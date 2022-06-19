package viewer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func (v *Viewer) GetTokenMetaForFileName(fileName string) (*fstore.FirestoreTokenMeta, error) {

	plaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		return nil, err
	}

	for _, metaID := range plaque.Plaque.TokenMetaIDList {
		meta, err := v.ReadMetadata(metaID)
		if err != nil {
			return nil, err
		}
		// filename could be from media id or external url
		if meta.TokenMeta.MediaID == strings.TrimSuffix(fileName, filepath.Ext(fileName)) || filepath.Base(meta.TokenMeta.ExternalMediaURL) == fileName {
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
		logger.Printf("loadPlaqueData failed to retrieve remote plaque, using local data. Error: %+v", err)
		// if we are offline, just return local plaque
		return localPlaque, nil
	}

	// if equal we do nothing, just return plaque
	if reflect.DeepEqual(remotePlaque, localPlaque) {
		return localPlaque, nil
	}

	// if not equal we overwrite local file with remote data
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
	defer jsonFile.Close()

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
// will not return tokens if offline and the token is not found on local OR token is not found on remote
// can only error if there are problems marshalling/writing json file which is unlikely
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
	defer jsonFile.Close()

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
// first will try to load from archive, then from external sources
func (v *Viewer) loadMedia(ctx context.Context, metas []*fstore.FirestoreTokenMeta) []*fstore.FirestoreTokenMeta {
	validMetas := make([]*fstore.FirestoreTokenMeta, 0, len(metas))
	for _, meta := range metas {
		if meta.TokenMeta.MediaID != "" {
			err := v.MediaClient.DownloadFileFromArchive(meta.MediaFileName())
			if err != nil {
				logger.Printf("loadMedia error - failed to load archive media for token %s with file name %s", meta.DocumentID, meta.MediaFileName())
				continue
			}
		} else if meta.TokenMeta.ExternalMediaURL != "" {
			err := v.MediaClient.DownloadFileFromURL(meta.TokenMeta.ExternalMediaURL)
			if err != nil {
				logger.Printf("loadMedia error - failed to load external media for token %s with url path %s", meta.DocumentID, meta.TokenMeta.ExternalMediaURL)
				continue
			}
		} else {
			logger.Printf("loadMedia error - token has no valid media links %s", meta.DocumentID)
			continue
		}

		// meta is valid if above media was found without error
		validMetas = append(validMetas, meta)
	}
	return validMetas
}

// getValidTokens returns list of token metas that both exist locally and have a local media file
func (v *Viewer) getValidTokens(tokenMetaIDList []string) []string {
	validTokenList := make([]string, 0, len(tokenMetaIDList))
	for _, tokenMetaID := range tokenMetaIDList {
		tokenMeta, err := v.ReadMetadata(tokenMetaID)
		if err != nil {
			logger.Printf("getValidTokens invalid token metadata for token %s", tokenMetaID)
			continue
		}

		localPath := filepath.Join(v.MediaDir, tokenMeta.MediaFileName())
		exists, err := storage.FileExists(localPath)
		if err != nil {
			logger.Printf("getValidTokens failed to check local file %s for metadata %s with error %v", localPath, tokenMetaID, err)
			continue
		}

		if !exists {
			logger.Printf("getValidTokens failed to find local file %s for metadata %s", localPath, tokenMetaID)
			continue
		}
		validTokenList = append(validTokenList, tokenMeta.DocumentID)
	}
	return validTokenList
}
