package viewer

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
)

func TestViewer(t *testing.T) {
	a := assert.New(t)
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config_test.json")

	testConfig := ViewerConfig{DocumentID: "1"}
	file, err := json.Marshal(testConfig)
	a.NoError(err)

	a.NoError(ioutil.WriteFile(configPath, file, 0644))

	v := Viewer{
		ConfigFile: configPath,
	}

	config, err := v.readConfig()
	a.NoError(err)

	a.Equal(testConfig.DocumentID, config.DocumentID)
}
