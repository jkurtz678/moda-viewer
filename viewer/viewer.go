package viewer

import (
	"context"
	"fmt"
	"jkurtz678/moda-viewer/storage"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var logger = log.New(os.Stdout, "[viewer] - ", log.Ldate|log.Ltime|log.Lshortfile)

// Viewer is an object that displays media and plaque information
type Viewer struct {
	PlaqueFile   string
	PlaylistFile string
	MediaDir     string
	MetadataDir  string
	DBClient
	MediaClient
	VideoPlayer
	PlaqueManager
	activeTokenMetaID string
	activeTokenLock   sync.Mutex
}

// NewViewer returns a new initialized viewer
func NewViewer(dbClient DBClient, storageClient *storage.FirebaseStorageClient) *Viewer {
	return &Viewer{
		PlaqueFile:    "plaque.json",
		PlaylistFile:  "playlist.m3u",
		MediaDir:      "media",
		MetadataDir:   "metadata",
		DBClient:      dbClient,
		MediaClient:   storageClient,
		VideoPlayer:   &VLCPlayer{},
		PlaqueManager: &PythonWebview{},
	}
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Start() error {
	logger.Printf("Start()")

	logger.Printf("loading plaque data...")
	plaque, err := v.loadPlaqueData(context.Background())
	if err != nil {
		return err
	}

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

	// write v3 playlist file
	m3uFile := ""
	for _, m := range metas {
		m3uFile += fmt.Sprintf("%s\n", filepath.Join(v.MediaDir, m.MediaFileName()))
	}
	err = os.WriteFile(v.PlaylistFile, []byte(m3uFile), 0644)
	if err != nil {
		return err
	}

	go func() {
		err = v.playMedia(v.PlaylistFile, func(mediaID string) {
			logger.Printf("playing media id: %s", mediaID)
			meta, err := v.GetTokenMetaForMediaID(mediaID)
			if err != nil {
				logger.Printf("error getting token meta in callback %s", err)
				return
			}

			//v.navigateURL(meta.DocumentID)

			v.activeTokenLock.Lock()
			v.activeTokenMetaID = meta.DocumentID
			v.activeTokenLock.Unlock()
		})
		if err != nil {
			logger.Printf("playMedia error %v", err)
		}
	}()
	time.Sleep(3 * time.Second)
	v.initPlaque()

	/* err = v.showPlaque()
	if err != nil {
		return err
	} */
	return nil
}
