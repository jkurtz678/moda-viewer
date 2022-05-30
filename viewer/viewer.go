package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/videoplayer"
	"jkurtz678/moda-viewer/webview"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var logger = log.New(os.Stdout, "[viewer] - ", log.Ldate|log.Ltime|log.Lshortfile)

// Viewer is an object that displays media and plaque information
type Viewer struct {
	PlaqueFile   string
	PlaylistFile string
	MediaDir     string
	MetadataDir  string
	fstore.DBClient
	MediaClient
	videoplayer.VideoPlayer
	webview.PlaqueManager
	Active bool // plaque has started up and is playing media
}

// NewViewer returns a new initialized viewer
func NewViewer(dbClient fstore.DBClient, storageClient *storage.FirebaseStorageClient) *Viewer {
	return &Viewer{
		PlaqueFile:    "plaque.json",
		PlaylistFile:  "playlist.m3u",
		MediaDir:      "media",
		MetadataDir:   "metadata",
		DBClient:      dbClient,
		MediaClient:   storageClient,
		VideoPlayer:   &videoplayer.VLCPlayer{},
		PlaqueManager: &webview.PythonWebview{},
	}
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Startup() error {
	logger.Printf("Startup()")

	// init plaque and player processes on their own threads
	go v.PlaqueManager.InitPlaque()
	go v.VideoPlayer.InitPlayer()

	logger.Printf("loading plaque data...")
	plaque, err := v.loadPlaqueData(context.Background())
	if err != nil {
		return err
	}

	err = v.ListenForPlaqueChanges(plaque)
	if err != nil {
		return err
	}

	return nil

}

// LoadAndPlayTokens accepts a plaque, loads its associated media/metadata, and tells the video player to start playing this media
func (v *Viewer) LoadAndPlayTokens(plaque *fstore.FirestorePlaque) error {
	if len(plaque.Plaque.TokenMetaIDList) == 0 {
		return fmt.Errorf("plaque has no assigned media")
	}

	logger.Printf("loading token metas...")
	metas, err := v.loadTokenMetas(context.Background(), plaque)
	if err != nil {
		return err
	}

	logger.Printf("loading media...")
	err = v.loadMedia(context.Background(), metas)
	if err != nil {
		return err
	}

	logger.Printf("playing media...")
	filepaths := make([]string, 0, len(metas))
	for _, m := range metas {
		filepaths = append(filepaths, url.QueryEscape(filepath.Join(v.MediaDir, m.MediaFileName())))
	}
	err = v.VideoPlayer.PlayFiles(filepaths)
	if err != nil {
		return err
	}
	return nil
}

func (v *Viewer) ListenForPlaqueChanges(plaque *fstore.FirestorePlaque) error {
	logger.Printf("ListenForPlaqueChanges - listening to changes for plaque: %s", plaque.DocumentID)
	err := v.DBClient.ListenPlaque(context.Background(), plaque.DocumentID, func(remotePlaque *fstore.FirestorePlaque) error {
		// update local plaque file with changes
		plaqueBytes, err := json.Marshal(remotePlaque)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(v.PlaqueFile, plaqueBytes, 0644)
		if err != nil {
			return err
		}

		err = v.LoadAndPlayTokens(remotePlaque)
		if err != nil {
			return err
		}
		v.Active = true
		return nil
	})
	if err != nil {
		logger.Printf("ListenForPlaqueChanges - listen error %v, retrying connection in 1 minute", err)
		// if plaque is not currently playing media and is offline, start with local plaque data
		if !v.Active {
			err = v.LoadAndPlayTokens(plaque)
			if err != nil {
				return err
			}
			v.Active = true
		}
		time.Sleep(1 * time.Minute)
		return v.ListenForPlaqueChanges(plaque)
	}
	return nil
}
