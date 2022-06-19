package storage

import (
	"log"
	"os"
	"path/filepath"
)

type FirebaseStorageClientStub struct {
	MediaDir string
}

// DownloadFileFromArchive creates empty file for tests
func (sc *FirebaseStorageClientStub) DownloadFileFromArchive(fileURI string) error {
	logger.Printf("FirebaseStorageClientStub.DownloadFileFromArchive - %s", fileURI)
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

// DownloadFileFromURL creates empty file for tests
func (sc *FirebaseStorageClientStub) DownloadFileFromURL(fileURL string) error {
	logger.Printf("FirebaseStorageClientStub.DownloadFileFromURL - %s", fileURL)
	f, err := os.Create(filepath.Join(sc.MediaDir, filepath.Base(fileURL)))
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
