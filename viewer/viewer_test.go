package viewer

import (
	"encoding/json"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"path/filepath"
	"testing"

	"github.com/franela/goblin"
	"github.com/google/go-cmp/cmp"
)

func TestViewer(t *testing.T) {
	g := goblin.Goblin(t)

	tmpdir := t.TempDir()

	g.Describe("Viewer (offline)", func() {
		// setup test config file
		testPlaque := fstore.FirestorePlaque{DocumentID: "1", Plaque: fstore.Plaque{Name: "test"}}
		file, err := json.Marshal(testPlaque)
		g.Assert(err).IsNil()
		configPath := filepath.Join(tmpdir, "config_test.json")
		g.Assert(ioutil.WriteFile(configPath, file, 0644)).IsNil()

		// setup test meta file
		testMeta := &fstore.FirestoreTokenMeta{DocumentID: "1", TokenMeta: fstore.TokenMeta{
			Name:   "starry night",
			Artist: "van gogh",
		}}
		file, err = json.Marshal(testMeta)
		g.Assert(err).IsNil()
		metaPath := filepath.Join(tmpdir, "1.json")
		g.Assert(ioutil.WriteFile(metaPath, file, 0644)).IsNil()

		// create viewer with stubbed VideoPlayer and PlaqueManager
		playerStub := &VideoPlayerStub{}
		plaqueStub := &PlaqueManagerStub{}
		fstoreClientStub := &fstore.FstoreClientStub{}

		v := Viewer{
			PlaqueFile:    configPath,
			MetadataDir:   tmpdir,
			MediaDir:      tmpdir,
			DBClient:      fstoreClientStub,
			VideoPlayer:   playerStub,
			PlaqueManager: plaqueStub,
		}

		g.It("Should load document id from config file", func() {
			plaque, err := v.readLocalPlaqueFile()
			g.Assert(err).IsNil()

			g.Assert(plaque.DocumentID).Equal(testPlaque.DocumentID)
			g.Assert(plaque.Plaque.Name).Equal(testPlaque.Plaque.Name)
		})

		g.It("Should properly parse metadata file", func() {
			retMeta, err := v.readMetadata(testMeta.DocumentID)
			g.Assert(err).IsNil()

			g.Assert(retMeta.DocumentID).Equal(testMeta.DocumentID)
			g.Assert(retMeta.TokenMeta.Name).Equal(testMeta.TokenMeta.Name)
			g.Assert(retMeta.TokenMeta.Artist).Equal(testMeta.TokenMeta.Artist)
			g.Assert(cmp.Equal(*retMeta, *testMeta)).IsTrue()
		})

		g.It("Should call video player and plaque manager with proper params", func() {
			playerStub.wg.Add(1) // ready player stub wait group
			g.Assert(v.Start()).IsNil()

			playerStub.wg.Wait() // block until player goroutine finishes

			g.Assert(playerStub.filepathPlayed).Equal(filepath.Join(tmpdir, "1.mp4"))
			g.Assert(cmp.Equal(*plaqueStub.tokenDisplayed, *testMeta)).IsTrue()
		})
	})

}
