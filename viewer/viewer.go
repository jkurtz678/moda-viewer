package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/videoplayer"
	"jkurtz678/moda-viewer/webview"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var logger = log.New(os.Stdout, "[viewer] - ", log.Ldate|log.Ltime|log.Lshortfile)

// Viewer is an object that displays media and plaque information
type Viewer struct {
	PlaqueFile  string
	MediaDir    string
	MetadataDir string
	fstore.DBClient
	storage.MediaClient
	videoplayer.VideoPlayer
	webview.PlaqueManager
	TestMode bool // plaque will not block and listen for changes, instead will close after playing media
	State    ViewerState

	stateLock sync.Mutex // lock for loading and loadErr values
	loading   bool       // boolean set to true when viewer is actively loading data
	loadErr   error      // error which viewer ran into while loading data, if any value is found here the viewer is considered in ViewerStateError
}

// NewViewer returns a new initialized viewer
func NewViewer(dbClient fstore.DBClient, storageClient *storage.FirebaseStorageClient) *Viewer {
	return &Viewer{
		PlaqueFile:    "plaque.json",
		MediaDir:      "media",
		MetadataDir:   "metadata",
		DBClient:      dbClient,
		MediaClient:   storageClient,
		VideoPlayer:   videoplayer.NewVLCPlayer(),
		PlaqueManager: &webview.PythonWebview{},
	}
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Startup() error {
	logger.Printf("Startup()")

	// start loading indicator
	v.stateLock.Lock()
	v.loading = true
	v.stateLock.Unlock()

	// init plaque and player processes on their own threads
	go v.PlaqueManager.InitPlaque()
	go v.VideoPlayer.InitPlayer()

	// pause to let plaque and player start up
	time.Sleep(time.Second)

	logger.Printf("loading plaque data...")
	plaque, err := v.loadPlaqueData(context.Background())
	// loadPlaqueData should only error if no local plaque is found (first start) and cannot connect to remote (no wifi)
	if err != nil {
		logger.Printf("Startup loadPlaqueData error - %+v", err)
		return err
	}

	// try to start plaque, if failure, wait 5 seconds and try again
	for {
		err = v.LoadAndPlayTokens(plaque)
		if err != nil {
			logger.Printf("LoadAndPlayTokens error %v - retrying in 5 seconds....", err)
			v.stateLock.Lock()
			v.loadErr = err
			v.stateLock.Unlock()
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	// if we make it here, plaque is now playing media
	v.stateLock.Lock()
	v.loading = false
	v.stateLock.Unlock()

	// if in test mode, exit after starting playback instead of listening
	if v.TestMode {
		return nil
	}

	// now listen for plaque changes on remote, should block perpetually
	v.ListenForPlaqueChanges(plaque)

	// if we get here plaque will exit, but there should be no way to move past ListenForPlaqueChanges
	return nil
}

// ListenForPlaqueChanges will trigger
func (v *Viewer) ListenForPlaqueChanges(plaque *fstore.FirestorePlaque) {
	startListenTime := time.Now()
	logger.Printf("ListenForPlaqueChanges - listening to changes for plaque: %s", plaque.DocumentID)
	err := v.DBClient.ListenPlaque(context.Background(), plaque.DocumentID, func(remotePlaque *fstore.FirestorePlaque) error {
		// if ListenPlaque has internet, it will immediately trigger the callback when first run we don't want, since the plaque is already running
		// To fix this we skip this callback if it has been called within 5 seconds of listen start
		if startListenTime.Add(time.Second * 5).After(time.Now()) {
			return nil
		}
		v.stateLock.Lock()
		v.loading = true
		v.loadErr = nil
		v.stateLock.Unlock()

		// update local plaque file with changes and play new tokens
		err := func() error {
			plaqueBytes, err := json.Marshal(remotePlaque)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(v.PlaqueFile, plaqueBytes, 0644)
			if err != nil {
				return err
			}

			err = v.LoadAndPlayTokens(remotePlaque)
			if err != nil {
				return err
			}
			return nil
		}()

		v.stateLock.Lock()
		v.loading = false
		if err != nil {
			logger.Printf("ListenForPlaqueChanges error %v", err)
			v.loadErr = err
		}
		v.stateLock.Unlock()

		return nil
	})
	// callback will not error, but possible that startup of listener will error, retry in 1 minute
	if err != nil {
		logger.Printf("ListenForPlaqueChanges - listen error %v, retrying connection in 1 minute", err)
		v.stateLock.Lock()
		v.loadErr = err
		v.stateLock.Unlock()
		time.Sleep(1 * time.Minute)
		v.ListenForPlaqueChanges(plaque)
	}
}

// LoadAndPlayTokens accepts a plaque, loads its associated media/metadata, and tells the video player to start playing this media
// possible errors:
// - PlayFiles error for logo is unlikely since it will block and retry until vlc is found (infinite loop here is more likely)
// - loadTokenMetas error is unlikely, only occurs on malformed token meta json
// - loadMedia error will only occur if all tokens in playlist fail to download, if only some fail it will be logged and these will be skipped
// - PlayFiles final now should only fail if vlc is not running, as tokens now have a verified local file that exists
func (v *Viewer) LoadAndPlayTokens(plaque *fstore.FirestorePlaque) error {
	logger.Printf("LoadAndPlayTokens called")

	// show moda logo during any file loading, and if errors are hit
	err := v.VideoPlayer.PlayFiles([]string{"moda-logo.png"})
	if err != nil {
		return err
	}

	// show moda logo if account_id is not set or no assigned tokens
	if plaque.Plaque.WalletAddress == "" {
		logger.Printf("LoadAndPlayTokens no connected user, showing logo")
		return nil
	}

	// show moda logo if no tokens are assigned to plaque
	if len(plaque.Plaque.TokenMetaIDList) == 0 {
		logger.Printf("LoadAndPlayTokens plaque has %v tokens and 0 valid tokens, showing logo", len(plaque.Plaque.TokenMetaIDList))
		return nil
	}

	logger.Printf("LoadAndPlayTokens loading token metas...")
	metas, err := v.loadTokenMetas(context.Background(), plaque)
	if err != nil {
		return err
	}

	logger.Printf("LoadAndPlayTokens loading media for %v metas", len(metas))
	// validTokenMetas are metas with associated media file that has been downloaded and exists locally
	validTokenMetas := v.loadMedia(context.Background(), metas)
	if len(validTokenMetas) == 0 {
		return fmt.Errorf("Viewer.loadMedia error - no valid tokens in list")
	}

	// log if any invalid tokens were found
	if len(validTokenMetas) != len(plaque.Plaque.TokenMetaIDList) {
		logger.Printf("LoadAndPlayTokens invalid tokens found - plaque has %v tokens and %v valid tokens, playing valid token(s)", len(plaque.Plaque.TokenMetaIDList), len(validTokenMetas))
	}

	logger.Printf("LoadAndPlayTokens playing media playlist of %v tokens", len(validTokenMetas))
	filepaths := make([]string, 0, len(validTokenMetas))
	for _, m := range validTokenMetas {
		filepaths = append(filepaths, url.QueryEscape(filepath.Join(v.MediaDir, m.MediaFileName())))
	}
	err = v.VideoPlayer.PlayFiles(filepaths)
	if err != nil {
		return err
	}
	return nil
}
