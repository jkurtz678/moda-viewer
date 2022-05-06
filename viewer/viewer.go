package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"cloud.google.com/go/firestore"
)

var logger = log.New(os.Stdout, "[viewer] - ", log.Ldate|log.Ltime|log.Lshortfile)

type DBClient interface {
	CreatePlaque(ctx context.Context, plaque *fstore.Plaque) (*fstore.FirestorePlaque, error)
	GetPlaque(ctx context.Context, documentID string) (*fstore.FirestorePlaque, error)
	UpdatePlaque(ctx context.Context, documentID string, update []firestore.Update) error

	CreateTokenMeta(ctx context.Context, tokenMeta *fstore.TokenMeta) (*fstore.FirestoreTokenMeta, error)
	GetTokenMeta(ctx context.Context, documentID string) (*fstore.FirestoreTokenMeta, error)
	GetTokenMetaList(ctx context.Context, documentIDList []string) ([]*fstore.FirestoreTokenMeta, error)
	UpdateTokenMeta(ctx context.Context, documentID string, update []firestore.Update) error
}

// MediaClient contains methods for managing media files videos/images/gifs
type MediaClient interface {
	DownloadFile(fileURI string) error
}

// Viewer is an object that displays media and plaque information
type Viewer struct {
	PlaqueFile  string
	MediaDir    string
	MetadataDir string
	PlaylistDir string
	DBClient
	MediaClient
	VideoPlayer
	PlaqueManager
}

// NewViewer returns a new initialized viewer
func NewViewer(dbClient DBClient, storageClient *storage.FirebaseStorageClient) *Viewer {
	return &Viewer{
		PlaqueFile:    "plaque.json",
		MediaDir:      "media",
		MetadataDir:   "metadata",
		PlaylistDir:   "playlist",
		DBClient:      dbClient,
		MediaClient:   storageClient,
		VideoPlayer:   &VLCPlayer{},
		PlaqueManager: &Webview{},
	}
}

// Start will play media and show the plaque as specified by the config file
func (v *Viewer) Start() error {
	logger.Printf("Start()")

	logger.Printf("loading plaque data...")
	plaque, err := v.loadPlaqueData(context.Background())
	if err != nil {
		return err
	}

	if len(plaque.Plaque.TokenMetaIDList) == 0 {
		return fmt.Errorf("plaque has no assigned media")
	}

	logger.Printf("loading token metas...")
	metas, err := v.loadTokenMetas(context.Background(), plaque)
	if err != nil {
		return err
	}

	logger.Printf("loading media...")
	err = v.loadMedia(context.Background(), metas)
	if err != nil {
		return err
	}

	// write v3 playlist file
	m3uFile := ""
	for _, m := range metas {
		m3uFile += fmt.Sprintf("%s\n", filepath.Join(v.MediaDir, m.MediaFileName()))
	}
	err = os.WriteFile("playlist.m3u", []byte(m3uFile), 0644)
	if err != nil {
		return err
	}

	v.initPlaque()

	go func() {
		err = v.playMedia("playlist.m3u", func(mediaID string) {
			logger.Printf("playing media id: %s", mediaID)
			meta, err := v.GetTokenMetaForMediaID(mediaID)
			if err != nil {
				log.Fatal(err)
			}
			v.navigateURL(meta.DocumentID)
			err = os.Truncate("vlc.txt", 100)
			if err != nil {
				log.Fatal(err)
			}
		})
		if err != nil {
			logger.Printf("playMedia error %v", err)
		}
	}()

	err = v.showPlaque()
	if err != nil {
		return err
	}
	return nil
}

func (v *Viewer) GetActiveTokenMeta() (*fstore.FirestoreTokenMeta, error) {
	plaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		return nil, err
	}

	meta, err := v.ReadMetadata(plaque.Plaque.TokenMetaIDList[0])
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (v *Viewer) GetTokenMetaForMediaID(mediaID string) (*fstore.FirestoreTokenMeta, error) {
	plaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		return nil, err
	}

	for _, metaID := range plaque.Plaque.TokenMetaIDList {
		meta, err := v.ReadMetadata(metaID)
		if err != nil {
			return nil, err
		}
		if meta.TokenMeta.MediaID == mediaID {
			return meta, nil
		}
	}
	return nil, fmt.Errorf("local meta file not found")
}

// loadPlaqueData loads the most up to date version of the plaque for this viewer, checking with the remote firestore
// - first loads the local plaque file
// - if local file does not exist, create on the remote
// - if does exist, retrieve corresponding plaque from firestore
// - if fail to retrieve from firestore, return local plaque
// - if firestore plaque found, compare to local plaque, if same return them
// - if not matching local plaque, overwrite and return remote
func (v *Viewer) loadPlaqueData(ctx context.Context) (*fstore.FirestorePlaque, error) {

	// read local file to get document id
	localPlaque, err := v.ReadLocalPlaqueFile()

	// if we cannot find a local plaque file, create one on the remote server
	if err != nil {
		logger.Printf("loadPlaqueData error reading local file: %+v, creating new plaque", err)
		remotePlaque, err := v.DBClient.CreatePlaque(ctx, new(fstore.Plaque))
		if err != nil {
			logger.Printf("loadPlaqueData failed to create new plaque, exiting with error: %+v", err)
			return nil, err
		}

		plaqueBytes, err := json.Marshal(remotePlaque)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(v.PlaqueFile, plaqueBytes, 0644)
		if err != nil {
			return nil, err
		}

		return remotePlaque, nil
	}

	// retrieve remote plaque that matches local document id
	remotePlaque, err := v.DBClient.GetPlaque(ctx, localPlaque.DocumentID)
	if err != nil {
		logger.Printf("loadPlaqueData failed to retrieve remote plaque: %+v", err)
		// if we are offline, just return local plaque
		return localPlaque, nil
	}

	if reflect.DeepEqual(remotePlaque, localPlaque) {
		return localPlaque, nil
	}

	// if not equal we overwrite file with remote data
	plaqueBytes, err := json.Marshal(remotePlaque)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(v.PlaqueFile, plaqueBytes, 0644)
	if err != nil {
		return nil, err
	}
	return remotePlaque, nil

}

// ReadLocalPlaqueFile attempts to read local plaque file
func (v *Viewer) ReadLocalPlaqueFile() (*fstore.FirestorePlaque, error) {
	jsonFile, err := os.Open(v.PlaqueFile)
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var plaque fstore.FirestorePlaque
	err = json.Unmarshal([]byte(byteValue), &plaque)
	if err != nil {
		return nil, err
	}

	return &plaque, err
}

// loadTokenMetas returns a list of token metas, loading from remote if online, returning local files if offline
func (v *Viewer) loadTokenMetas(ctx context.Context, plaque *fstore.FirestorePlaque) ([]*fstore.FirestoreTokenMeta, error) {
	localMetas := make([]*fstore.FirestoreTokenMeta, 0)
	for _, docID := range plaque.Plaque.TokenMetaIDList {
		fToken, err := v.ReadMetadata(docID)
		if err != nil {
			// if err, assume the metadata file has not been loaded locally yet
			continue
		}
		localMetas = append(localMetas, fToken)
	}

	remoteMetas, err := v.DBClient.GetTokenMetaList(ctx, plaque.Plaque.TokenMetaIDList)
	if err != nil {
		// if offline, return local tokens
		logger.Printf("loadTokenMetas GetTokenMetaList error: %+v", err)
		return localMetas, nil
	}

	// put local tokens in a map so we can ignore order
	localMetaMap := make(map[string]*fstore.FirestoreTokenMeta, len(localMetas))
	for _, meta := range localMetas {
		localMetaMap[meta.DocumentID] = meta
	}

	// if local token does not exist or match remote token, overwrite local file
	for _, meta := range remoteMetas {
		if reflect.DeepEqual(localMetaMap[meta.DocumentID], meta) {
			continue
		}

		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return nil, err
		}

		fileName := fmt.Sprintf("%s.json", meta.DocumentID)
		filePath := filepath.Join(v.MetadataDir, fileName)
		err = ioutil.WriteFile(filePath, metaBytes, 0644)
		if err != nil {
			return nil, err
		}
	}

	return remoteMetas, nil

}

// ReadMetadata reads and returns the metadata file for the given document id
func (v *Viewer) ReadMetadata(documentID string) (*fstore.FirestoreTokenMeta, error) {
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

// loadMedia will download all media for metas, ensuring that all media files are ready for playback
func (v *Viewer) loadMedia(ctx context.Context, metas []*fstore.FirestoreTokenMeta) error {
	for _, meta := range metas {
		err := v.MediaClient.DownloadFile(meta.MediaFileName())
		if err != nil {
			return err
		}
	}
	return nil
}
