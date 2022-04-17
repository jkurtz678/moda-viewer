package viewer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"log"
	"os"
	"path/filepath"
)

var logger = log.New(os.Stdout, "[viewer] - ", log.Ldate|log.Ltime|log.Lshortfile)

// Viewer is an object that displays media and plaque information
type Viewer struct {
	ConfigFile  string
	MediaDir    string
	MetadataDir string
	PlaylistDir string
	VideoPlayer
	PlaqueManager
}

// NewViewer returns a new initialized viewer
func NewViewer() *Viewer {
	return &Viewer{
		ConfigFile:    "config.json",
		MediaDir:      "media",
		MetadataDir:   "metadata",
		PlaylistDir:   "playlist",
		VideoPlayer:   &VLCPlayer{},
		PlaqueManager: &Webview{},
	}
}

type ViewerConfig struct {
	DocumentID string `json:"document_id"`
	Playlist   bool   `json:"playlist"`
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Start() error {
	logger.Printf("Start()")

	meta, err := v.GetActiveTokenMeta()
	if err != nil {
		return err
	}

	go func() {
		mediaPath := filepath.Join(v.MediaDir, fmt.Sprintf("%s.mp4", meta.DocumentID))
		err = v.playMedia(mediaPath)
		if err != nil {
			logger.Printf("playMedia error %v", err)
		}
	}()

	err = v.showPlaque(meta)
	if err != nil {
		return err
	}
	return nil
}

func (v *Viewer) GetActiveTokenMeta() (*fstore.FirestoreTokenMeta, error) {
	config, err := v.readConfig()
	if err != nil {
		return nil, err
	}

	if config.Playlist {
		return nil, fmt.Errorf("error - playlists not implemented")
	}

	meta, err := v.readMetadata(config.DocumentID)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// readConfig reads and returns the config file for the viewer
func (v *Viewer) readConfig() (*ViewerConfig, error) {
	jsonFile, err := os.Open(v.ConfigFile)
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var config ViewerConfig
	err = json.Unmarshal([]byte(byteValue), &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}

// readMetadata reads and returns the metadata file for the given document id
func (v *Viewer) readMetadata(documentID string) (*fstore.FirestoreTokenMeta, error) {
	fileName := fmt.Sprintf("%s.json", documentID)
	jsonFile, err := os.Open(filepath.Join(v.MetadataDir, fileName))
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var meta fstore.FirestoreTokenMeta
	err = json.Unmarshal([]byte(byteValue), &meta)
	if err != nil {
		return nil, err
	}

	return &meta, err
}
