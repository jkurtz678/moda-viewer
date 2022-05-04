package storage

import (
	"log"
	"os"
	"path/filepath"
)

type FirebaseStorageClientStub struct {
	MediaDir string
}

// DownloadFile creates empty file for tests
func (sc *FirebaseStorageClientStub) DownloadFile(fileURI string) error {
	logger.Printf("FirebaseStorageClientStub.DownloadFile - %s", fileURI)
	f, err := os.Create(filepath.Join(sc.MediaDir, fileURI))
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
