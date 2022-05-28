package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"os"
	"path/filepath"
	"reflect"

	"cloud.google.com/go/firestore"
)

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
		logger.Printf("updating local meta for token %s", meta.TokenMeta.Name)

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
