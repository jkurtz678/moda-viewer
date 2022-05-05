package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func (sc *FirebaseStorageClient) handleQueue() {
	for fileURI := range sc.downloadQueue {
		var err error
		if strings.Contains(fileURI, "https://") {
			err = sc.downloadFileFromURL(fileURI)
		} else {
			err = sc.DownloadFile(fileURI)
		}
		if err != nil {
			logger.Printf("error downloading file %+v", err)
		}
	}
}
func (sc *FirebaseStorageClient) downloadFileFromURL(fileURL string) error {
	logger.Printf("downloadFileFromURL - %s", fileURL)
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// create media dir if it does not exist, does nothing if already exists
	err = os.MkdirAll(sc.mediaDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("FirebaseStorageClient.performDownload - Failed to create media dir %s error %s", sc.mediaDir, err)
	}

	splitURL := strings.Split(fileURL, "/")
	fileName := splitURL[len(splitURL)-1]
	localPath := filepath.Join(sc.mediaDir, fileName)

	// Create the file
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (sc *FirebaseStorageClient) DownloadFile(fileURI string) error {
	localPath := filepath.Join(sc.mediaDir, fileURI)
	logger.Printf("downloadFileFromFirebase – %s", fileURI)

	exists, err := FileExists(localPath)
	if err != nil {
		return fmt.Errorf("FirebaseStorageClient.downloadFileFromFirebase - error checking file status %s", err)
	}
	if exists {
		log.Print("FirebaseStorageClient.downloadFileFromFirebase - File already exists, skipping download")
		return nil
	}

	data, err := sc.retrieveFileFromFirebase(fileURI)
	if err != nil {
		return fmt.Errorf("FirebaseStorageClient.downloadFileFromFirebase - retrieveFile %s error %s", fileURI, err)
	}

	log.Println("FirebaseStorageClient.downloadFileFromFirebase - Writing file...")

	// create media dir if it does not exist, does nothing if already exists
	err = os.MkdirAll(sc.mediaDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("FirebaseStorageClient.performDownload - Failed to create media dir %s error %s", sc.mediaDir, err)
	}

	err = os.WriteFile(localPath, data, 0644)
	if err != nil {
		return fmt.Errorf("FirebaseStorageClient.performDownload - WriteFile %s error %s", fileURI, err)
	}
	log.Printf("FirebaseStorageClient.downloadFileFromFirebase - download complete for file %s", localPath)
	return nil
}

// downloadFromCloudStorage will retrieve a file from firebase storage
func (sc *FirebaseStorageClient) retrieveFileFromFirebase(fileURI string) ([]byte, error) {
	config := &firebase.Config{
		StorageBucket: sc.storageBucketURL,
	}
	opt := option.WithCredentialsFile(sc.credentialsFile)
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		return nil, err
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return nil, err
	}
	log.Println("FirebaseStorageClient.retrieveFileFromFirebase - reading from bucket")

	rc, err := bucket.Object(fileURI).NewReader(context.Background())
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}
