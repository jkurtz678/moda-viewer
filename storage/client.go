package storage

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

var logger = log.New(os.Stdout, "[storage] - ", log.Ldate|log.Ltime|log.Lshortfile)

type FirebaseStorageClient struct {
	storageBucketURL string // url of firebase storage bucket
	credentialsFile  string // file path to firebase credentials
	mediaDir         string // path to directory where media files are stored
	downloadQueue    chan string
}

func NewFirebaseStorageClient(storageBucketURL, credentialsFile, mediaDir string) *FirebaseStorageClient {
	client := &FirebaseStorageClient{
		storageBucketURL: storageBucketURL,
		credentialsFile:  credentialsFile,
		mediaDir:         mediaDir,
		downloadQueue:    make(chan string, 10),
	}

	// start one worker who will wait on the downloadQueue
	go client.handleQueue()

	return client
}

// downloadFileFromFirebase will check if a file already exists locally, otherwise will download it from firebase
func (sc *FirebaseStorageClient) AttemptDownloadFromFirebase(fileURI string) (bool, error) {
	localPath := filepath.Join(sc.mediaDir, fileURI)
	logger.Printf("DownloadFileFromFirebase – attempting download %s", localPath)

	exists, err := FileExists(localPath)
	if err != nil {
		return false, err
	}
	if exists {
		logger.Println("File already exists, skipping download")
		return false, nil
	}

	sc.downloadQueue <- fileURI

	return true, nil
}

// AttemptDownloadFromURL will insert a url into the download queue
func (sc *FirebaseStorageClient) AttemptDownloadFromURL(url string) (bool, error) {
	splitURL := strings.Split(url, "/")
	fileName := splitURL[len(splitURL)-1]
	localPath := filepath.Join(sc.mediaDir, fileName)
	exists, err := FileExists(localPath)
	if err != nil {
		return false, err
	}
	if exists {
		logger.Println("File already exists, skipping download")
		return false, nil
	}

	sc.downloadQueue <- url

	return true, nil
}
