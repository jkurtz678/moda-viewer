package viewer

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/franela/goblin"
)

func TestViewer(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Viewer", func() {
		g.It("Should load document id from config file", func() {
			tmpdir := t.TempDir()
			configPath := filepath.Join(tmpdir, "config_test.json")

			testConfig := ViewerConfig{DocumentID: "1", Playlist: true}
			file, err := json.Marshal(testConfig)
			g.Assert(err).IsNil()

			g.Assert(ioutil.WriteFile(configPath, file, 0644)).IsNil()

			v := Viewer{
				ConfigFile: configPath,
			}

			config, err := v.readConfig()
			g.Assert(err).IsNil()

			g.Assert(config.DocumentID).Equal(testConfig.DocumentID)
			g.Assert(config.Playlist).Equal(testConfig.Playlist)
		})
	})
}
