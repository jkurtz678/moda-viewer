package viewer

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/videoplayer"
	"jkurtz678/moda-viewer/webview"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/franela/goblin"
	"github.com/google/go-cmp/cmp"
)

func TestViewer(t *testing.T) {
	ctx := context.Background()
	g := goblin.Goblin(t)

	g.Describe("Viewer (offline)", func() {
		tmpdir := t.TempDir()
		// setup test config file
		testPlaque := fstore.FirestorePlaque{
			DocumentID: "1",
			Plaque: fstore.Plaque{Name: "test",
				WalletAddress:   "test",
				TokenMetaIDList: []string{"m1", "m2"},
			},
		}
		file, err := json.Marshal(testPlaque)
		g.Assert(err).IsNil()
		configPath := filepath.Join(tmpdir, "config_test.json")
		g.Assert(ioutil.WriteFile(configPath, file, 0644)).IsNil()

		// setup 2 test meta files
		testMeta1 := &fstore.FirestoreTokenMeta{
			DocumentID: "m1",
			TokenMeta: fstore.TokenMeta{
				Name:      "starry night",
				Artist:    "van gogh",
				MediaID:   "s1",
				MediaType: ".mp4",
			}}
		file, err = json.Marshal(testMeta1)
		g.Assert(err).IsNil()
		metaPath := filepath.Join(tmpdir, "m1.json")
		g.Assert(ioutil.WriteFile(metaPath, file, 0644)).IsNil()

		testMeta2 := &fstore.FirestoreTokenMeta{
			DocumentID: "m2",
			TokenMeta: fstore.TokenMeta{
				Name:      "a sunday on the la grande jatte",
				Artist:    "seurat",
				MediaID:   "s2",
				MediaType: ".mp4",
			}}
		file, err = json.Marshal(testMeta2)
		g.Assert(err).IsNil()
		metaPath = filepath.Join(tmpdir, "m2.json")
		g.Assert(ioutil.WriteFile(metaPath, file, 0644)).IsNil()
		metas := []*fstore.FirestoreTokenMeta{testMeta1, testMeta2}

		// create empty media files
		_, err = os.Create(testMeta1.MediaFileName())
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Create(testMeta1.MediaFileName())
		if err != nil {
			t.Fatal(err)
		}

		// create viewer with stubbed VideoPlayer and PlaqueManager
		v := NewTestViewer(tmpdir)
		playerStub := &videoplayer.VideoPlayerStub{}
		plaqueStub := &webview.PlaqueManagerStub{}
		fstoreStub := &fstore.FstoreClientStub{}
		v.VideoPlayer = playerStub
		v.PlaqueManager = plaqueStub
		v.DBClient = fstoreStub

		g.It("Should load document id from config file", func() {
			plaque, err := v.ReadLocalPlaqueFile()
			g.Assert(err).IsNil()

			g.Assert(plaque.DocumentID).Equal(testPlaque.DocumentID)
			g.Assert(plaque.Plaque.Name).Equal(testPlaque.Plaque.Name)
		})

		g.It("Should properly parse metadata file", func() {
			retMeta, err := v.ReadMetadata(testMeta1.DocumentID)
			g.Assert(err).IsNil()

			g.Assert(retMeta.DocumentID).Equal(testMeta1.DocumentID)
			g.Assert(retMeta.TokenMeta.Name).Equal(testMeta1.TokenMeta.Name)
			g.Assert(retMeta.TokenMeta.Artist).Equal(testMeta1.TokenMeta.Artist)
			g.Assert(cmp.Equal(*retMeta, *testMeta1)).IsTrue()
		})

		g.It("Should call video player and plaque manager with proper params", func() {
			playerStub.PlayFilesWaitGroup.Add(1) // read fstore waitgroup

			// startup will block so start in own routine
			go func() {
				g.Assert(v.Startup()).IsNil()
			}()

			playerStub.PlayFilesWaitGroup.Wait() // block until viewer fully starts up

			// plaque and player have been started
			g.Assert(plaqueStub.PlaqueInit).IsTrue()
			g.Assert(playerStub.PlayerInit).IsTrue()

			// player is playing correct files
			filepaths := make([]string, 0)
			for _, m := range metas {
				filepaths = append(filepaths, url.QueryEscape(filepath.Join(v.MediaDir, m.MediaFileName())))
			}
			g.Assert(playerStub.ActivePlaylistFilepaths).Equal(filepaths)
		})
	})

	g.Describe("Viewer (fstore emulator)", func() {
		tmpdir := t.TempDir()
		v := NewTestViewer(tmpdir)
		v.DBClient = fstore.NewFirestoreTestClient(context.Background())
		playerStub := &videoplayer.VideoPlayerStub{}
		plaqueStub := &webview.PlaqueManagerStub{}
		v.VideoPlayer = playerStub
		v.PlaqueManager = plaqueStub
		v.TestMode = true // set test mode so viewer does not block

		g.It("should show moda logo if no account is assigned to plaque", func() {
			playerStub.PlayFilesWaitGroup.Add(1) // ready player stub wait group
			g.Assert(v.Startup()).IsNil()
			g.Assert(playerStub.ActivePlaylistFilepaths).Equal([]string{"moda-logo.png"})
		})

		g.It("should show moda logo no tokens are selected", func() {
			plaque, err := v.ReadLocalPlaqueFile()
			g.Assert(err).IsNil()
			err = v.DBClient.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{
				{Path: "wallet_address", Value: "test_account_id"},
			})
			g.Assert(err).IsNil()
			playerStub.PlayFilesWaitGroup.Add(1) // ready player stub wait group
			g.Assert(v.Startup()).IsNil()
			g.Assert(playerStub.ActivePlaylistFilepaths).Equal([]string{"moda-logo.png"})
		})

		g.It("should load metas and media from firebase after assigning media", func() {
			plaque, err := v.ReadLocalPlaqueFile()
			g.Assert(err).IsNil()

			meta1, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "starry night", MediaID: "s1", MediaType: ".png"})
			g.Assert(err).IsNil()

			meta2, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "a sunday on the la grande jatte", MediaID: "s2", MediaType: ".mp4"})
			g.Assert(err).IsNil()

			// add account id so that scanning screen is not shown
			// add token meta list so that plaque will play meta
			err = v.DBClient.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{
				{Path: "token_meta_id_list", Value: []string{meta1.DocumentID, meta2.DocumentID}},
			})
			g.Assert(err).IsNil()

			playerStub.PlayFilesWaitGroup.Add(1) // ready player stub wait group
			g.Assert(v.Startup()).IsNil()

			// ensure metas are loaded
			localMeta1, err := v.ReadMetadata(meta1.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta1.TokenMeta.MediaID).Equal(localMeta1.TokenMeta.MediaID)

			localMeta2, err := v.ReadMetadata(meta2.DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta2.TokenMeta.MediaID).Equal(localMeta2.TokenMeta.MediaID)

			metas := []*fstore.FirestoreTokenMeta{localMeta1, localMeta2}

			// ensure media files exist
			exists, err := storage.FileExists(filepath.Join(v.MediaDir, localMeta1.MediaFileName()))
			g.Assert(err).IsNil()
			g.Assert(exists).IsTrue()
			exists, err = storage.FileExists(filepath.Join(v.MediaDir, localMeta2.MediaFileName()))
			g.Assert(err).IsNil()
			g.Assert(exists).IsTrue()

			filepaths := make([]string, 0)
			for _, m := range metas {
				filepaths = append(filepaths, url.QueryEscape(filepath.Join(v.MediaDir, m.MediaFileName())))
			}

			g.Assert(playerStub.ActivePlaylistFilepaths).Equal(filepaths)
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

			localPlaque, err := v.ReadLocalPlaqueFile()
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
			localPlaque, err := v.ReadLocalPlaqueFile()
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
		err = v.DBClient.UpdatePlaque(ctx, plaque.DocumentID, []firestore.Update{{Path: "token_meta_id_list", Value: []string{meta1.DocumentID, meta2.DocumentID}}})
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
			localMeta1, err := v.ReadMetadata(metas[0].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta1.DocumentID).Equal(meta1.DocumentID)
			g.Assert(localMeta1.TokenMeta.Name).Equal(meta1.TokenMeta.Name)

			localMeta2, err := v.ReadMetadata(metas[1].DocumentID)
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
			localMeta1, err = v.ReadMetadata(metas[0].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta1.DocumentID).Equal(meta1.DocumentID)
			g.Assert(localMeta1.TokenMeta.Name).Equal("starry night update")

			localMeta2, err = v.ReadMetadata(metas[1].DocumentID)
			g.Assert(err).IsNil()
			g.Assert(localMeta2.DocumentID).Equal(meta2.DocumentID)
			g.Assert(localMeta2.TokenMeta.Name).Equal(meta2.TokenMeta.Name)
		})
	})
}

func NewTestViewer(tmpdir string) *Viewer {
	configPath := filepath.Join(tmpdir, "config_test.json")
	playerStub := &videoplayer.VideoPlayerStub{}
	plaqueStub := &webview.PlaqueManagerStub{}
	fstoreClientStub := &fstore.FstoreClientStub{}
	storageClientStub := &storage.FirebaseStorageClientStub{
		MediaDir: tmpdir,
	}
	return &Viewer{
		PlaqueFile:    configPath,
		MetadataDir:   tmpdir,
		MediaDir:      tmpdir,
		DBClient:      fstoreClientStub,
		MediaClient:   storageClientStub,
		VideoPlayer:   playerStub,
		PlaqueManager: plaqueStub,
	}
}

func TestCMDLogger(t *testing.T) {
	t.Skip()
	cmd := exec.Command("vlc", "../media/skate.mp4")
	stdout, _ := cmd.StdoutPipe()
	f, _ := os.Create("stdout.log")

	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(io.MultiWriter(f, os.Stdout), stdout)
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		t.Fatal(err)
	}

}
