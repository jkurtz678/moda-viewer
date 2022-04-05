package viewer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Viewer struct {
	ConfigFile string
}

type ViewerConfig struct {
	DocumentID string `json:"document_id"`
}

func (v *Viewer) ShowMedia() error {
	config, err := v.readConfig()
	if err != nil {
		return err
	}
	_ = config

	return nil
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
