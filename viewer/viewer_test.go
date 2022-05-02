package viewer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/franela/goblin"
	"github.com/google/go-cmp/cmp"
)

func NewTestViewer(tmpdir string) *Viewer {
	configPath := filepath.Join(tmpdir, "config_test.json")
	playerStub := &VideoPlayerStub{}
	plaqueStub := &PlaqueManagerStub{}
	fstoreClientStub := &fstore.FstoreClientStub{}
	return &Viewer{
		PlaqueFile:    configPath,
		MetadataDir:   tmpdir,
		MediaDir:      tmpdir,
		DBClient:      fstoreClientStub,
		VideoPlayer:   playerStub,
		PlaqueManager: plaqueStub,
	}
}

func TestViewer(t *testing.T) {
	ctx := context.Background()
	g := goblin.Goblin(t)

	g.Describe("Viewer", func() {
		tmpdir := t.TempDir()
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
		v := NewTestViewer(tmpdir)
		playerStub := &VideoPlayerStub{}
		plaqueStub := &PlaqueManagerStub{}
		v.VideoPlayer = playerStub
		v.PlaqueManager = plaqueStub

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

	g.Describe("loadPlaqueData (online)", func() {

		g.It("should create file if none exists", func() {
			tmpdir := t.TempDir()
			v := NewTestViewer(tmpdir)
			v.DBClient = fstore.NewFirestoreTestClient(context.Background())

			plaque, err := v.loadPlaqueData(context.Background())
			g.Assert(err).IsNil()
			g.Assert(plaque.DocumentID != "").IsTrue()
			g.Assert(plaque.Plaque.Name == "").IsTrue()

			// ensure plaque file was created
			_, err = os.Stat(v.PlaqueFile)
			g.Assert(err).IsNil()

			localPlaque, err := v.readLocalPlaqueFile()
			g.Assert(err).IsNil()
			g.Assert(localPlaque.DocumentID).Equal(plaque.DocumentID)

			// ensure it also exists on remote
			remotePlaque, err := v.DBClient.GetPlaque(ctx, plaque.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(remotePlaque.DocumentID).Equal(plaque.DocumentID)
		})

		g.It("should update local file if remote changes", func() {
			tmpdir := t.TempDir()
			v := NewTestViewer(tmpdir)
			v.DBClient = fstore.NewFirestoreTestClient(context.Background())

			plaque, err := v.loadPlaqueData(context.Background())
			g.Assert(err).IsNil()
			g.Assert(plaque.Plaque.Name).Equal("")
			// now local file should exist

			// change remote
			g.Assert(v.DBClient.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{
				{Path: "name", Value: "update-test"},
			})).IsNil()

			// loadPlaqueData should trigger overwriting of local file
			plaque2, err := v.loadPlaqueData(context.Background())
			g.Assert(err).IsNil()
			g.Assert(plaque2.DocumentID).Equal(plaque.DocumentID)
			g.Assert(plaque2.Plaque.Name).Equal("update-test")

			// ensure local file has been changed
			localPlaque, err := v.readLocalPlaqueFile()
			g.Assert(err).IsNil()
			g.Assert(localPlaque.DocumentID).Equal(plaque.DocumentID)
			g.Assert(localPlaque.Plaque.Name).Equal("update-test")

			// loading plaque data with no change should return the same plaque
			plaque3, err := v.loadPlaqueData(context.Background())
			g.Assert(err).IsNil()
			g.Assert(plaque3.DocumentID).Equal(plaque.DocumentID)
			g.Assert(plaque3.Plaque.Name).Equal("update-test")
		})
	})
	g.Describe("loadTokenMetas (online)", func() {
		tmpdir := t.TempDir()
		v := NewTestViewer(tmpdir)
		v.DBClient = fstore.NewFirestoreTestClient(context.Background())

		// setup plaque and metas
		plaque, err := v.loadPlaqueData(ctx)
		g.Assert(err).IsNil()

		meta1, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "starry night"})
		g.Assert(err).IsNil()

		meta2, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "a sunday on the la grande jatte"})
		g.Assert(err).IsNil()

		// add token to plaques
		err = v.DBClient.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{{Path: "token_meta_document_id_list", Value: []string{meta1.DocumentID, meta2.DocumentID}}})
		g.Assert(err).IsNil()

		// run load plaque data again to get updated values (and ensure local plaque is matching)
		plaque, err = v.loadPlaqueData(ctx)
		g.Assert(err).IsNil()

		g.It("should load and create local files for token metas", func() {
			metas, err := v.loadTokenMetas(ctx, plaque)
			g.Assert(err).IsNil()
			g.Assert(len(metas)).Equal(2)
			g.Assert(metas[0].DocumentID).Equal(meta1.DocumentID)
			g.Assert(metas[0].TokenMeta.Name).Equal(meta1.TokenMeta.Name)
			g.Assert(metas[1].DocumentID).Equal(meta2.DocumentID)
			g.Assert(metas[1].TokenMeta.Name).Equal(meta2.TokenMeta.Name)

			// now ensure local files match
			localMeta1, err := v.readMetadata(metas[0].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta1.DocumentID).Equal(meta1.DocumentID)
			g.Assert(localMeta1.TokenMeta.Name).Equal(meta1.TokenMeta.Name)

			localMeta2, err := v.readMetadata(metas[1].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta2.DocumentID).Equal(meta2.DocumentID)
			g.Assert(localMeta2.TokenMeta.Name).Equal(meta2.TokenMeta.Name)

			// update remote, ensure local files update
			g.Assert(v.DBClient.UpdateTokenMeta(ctx, meta1.DocumentID, []firestore.Update{{
				Path: "name", Value: "starry night update",
			}})).IsNil()
			metas, err = v.loadTokenMetas(ctx, plaque)
			g.Assert(err).IsNil()
			g.Assert(metas[0].DocumentID).Equal(meta1.DocumentID)
			g.Assert(metas[0].TokenMeta.Name).Equal("starry night update")
			g.Assert(metas[1].DocumentID).Equal(meta2.DocumentID)
			g.Assert(metas[1].TokenMeta.Name).Equal(meta2.TokenMeta.Name)

			// now ensure local files match
			localMeta1, err = v.readMetadata(metas[0].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta1.DocumentID).Equal(meta1.DocumentID)
			g.Assert(localMeta1.TokenMeta.Name).Equal("starry night update")

			localMeta2, err = v.readMetadata(metas[1].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta2.DocumentID).Equal(meta2.DocumentID)
			g.Assert(localMeta2.TokenMeta.Name).Equal(meta2.TokenMeta.Name)
		})
	})
}
