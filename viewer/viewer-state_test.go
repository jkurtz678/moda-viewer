package viewer

import (
	"context"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/videoplayer"
	"jkurtz678/moda-viewer/webview"
	"path/filepath"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

func TestViewerState(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	a := assert.New(t)

	type testCase struct {
		name     string
		setup    func(*Viewer)
		asserts  func(*Viewer, error)
		offline  bool // viewer is offline, cannot connect to remote firestore instance
		blocking bool // viewer blocks instead of exits automatically
	}
	testCases := []testCase{
		{
			name: "initial-start-offline", // player exits on startup, no wifi and no plaque
			asserts: func(v *Viewer, err error) {
				a.Error(err)
			},
			offline: true,
		},
		{
			name: "initial-start-online", // player creates plaque file and shows scan image
			asserts: func(v *Viewer, err error) {
				viewerStateData := v.GetViewerState()
				a.Equal(ViewerStateQrScan, viewerStateData.State)
				a.NotNil(viewerStateData.Plaque)
				a.Nil(viewerStateData.ActiveTokenMeta)
				a.NoError(err)
			},
			blocking: true,
		},
		{
			name: "no-tokens", // user connected to plaque but no tokens assigned
			setup: func(v *Viewer) {
				// load initial plaque file with address
				SetupTestPlaqueData(t, v, false)
			},
			asserts: func(v *Viewer, err error) {
				viewerStateData := v.GetViewerState()
				a.Equal(ViewerStateNoValidTokens, viewerStateData.State)
				a.NotNil(viewerStateData.Plaque)
				a.Nil(viewerStateData.ActiveTokenMeta)
				a.NoError(err)
			},
			blocking: true,
		},
		{
			name: "display-art", // user connected to plaque and playing tokens
			setup: func(v *Viewer) {
				SetupTestPlaqueData(t, v, true)
			},
			asserts: func(v *Viewer, err error) {
				viewerStateData := v.GetViewerState()
				a.Equal(ViewerStateDisplay, viewerStateData.State)
				a.NotNil(viewerStateData.Plaque)
				a.NotNil(viewerStateData.ActiveTokenMeta)
				a.NoError(err)
			},
			blocking: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tmpdir := t.TempDir()
			viewer, playerStub, plaqueStub := ViewerTestSetup(tmpdir)
			_ = playerStub
			_ = plaqueStub

			if test.offline {
				viewer.DBClient = &fstore.FstoreClientStub{}
			}

			if test.setup != nil {
				test.setup(viewer)
			}

			playerStub.PlayFilesWaitGroup.Add(1)
			if test.blocking {
				go func() {
					a.NoError(viewer.Startup())
				}()

				// wait for viewer to get to play vlc function and then run asserts
				playerStub.PlayFilesWaitGroup.Wait()
				test.asserts(viewer, nil)
				return
			}
			// otherwise viewer is expected to exit automatically
			err := viewer.Startup()
			test.asserts(viewer, err)

		})
	}
}

func ViewerTestSetup(tmpdir string) (*Viewer, *videoplayer.VideoPlayerStub, *webview.PlaqueManagerStub) {
	configPath := filepath.Join(tmpdir, "config_test.json")
	playerStub := &videoplayer.VideoPlayerStub{}
	plaqueStub := &webview.PlaqueManagerStub{}
	fstoreTestClient := fstore.NewFirestoreTestClient(context.Background())
	storageClientStub := &storage.FirebaseStorageClientStub{
		MediaDir: tmpdir,
	}
	v := &Viewer{
		PlaqueFile:    configPath,
		MetadataDir:   tmpdir,
		MediaDir:      tmpdir,
		DBClient:      fstoreTestClient,
		MediaClient:   storageClientStub,
		VideoPlayer:   playerStub,
		PlaqueManager: plaqueStub,
	}
	return v, playerStub, plaqueStub
}

func SetupTestPlaqueData(t *testing.T, v *Viewer, metas bool) {
	ctx := context.Background()
	p, err := v.loadPlaqueData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = v.DBClient.UpdatePlaque(ctx, p.DocumentID, []firestore.Update{
		{Path: "wallet_address", Value: "test_account"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if metas {
		meta1, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "starry night", MediaID: "s1", MediaType: ".png"})
		if err != nil {
			t.Fatal(err)
		}

		meta2, err := v.DBClient.CreateTokenMeta(ctx, &fstore.TokenMeta{Name: "a sunday on the la grande jatte", MediaID: "s2", MediaType: ".mp4"})
		if err != nil {
			t.Fatal(err)
		}
		err = v.DBClient.UpdatePlaque(ctx, p.DocumentID, []firestore.Update{{Path: "token_meta_id_list", Value: []string{meta1.DocumentID, meta2.DocumentID}}})
		if err != nil {
			t.Fatal(err)
		}
	}

}
