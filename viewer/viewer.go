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
		PlaylistDir:   "../playlist",
		VideoPlayer:   &VLCPlayer{},
		PlaqueManager: &Webview{},
	}
}

type ViewerConfig struct {
	DocumentID string `json:"document_id"`
	Playlist   bool   `json:"playlist"`
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Start() {
	logger.Printf("Start()")
	config, err := v.readConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if config.Playlist {
		logger.Fatal(fmt.Errorf("error - playlists not implemented"))
	}

	meta, err := v.readMetadata(config.DocumentID)
	if err != nil {
		logger.Fatal(err)
	}

	go func() {
		err := v.playMedia(meta.DocumentID)
		if err != nil {
			logger.Printf("playMedia error %v", err)
		}
	}()

	err = v.showPlaque(meta)
	if err != nil {
		logger.Fatal(err)
	}
}

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

func (v *Viewer) readMetadata(documentID string) (*fstore.FirestoreTokenMeta, error) {
	t, err := os.Getwd()
	logger.Printf("metadata pwd: %s, err %s", t, err)
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
