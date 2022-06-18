package main

import (
	"context"
	"fmt"
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	serviceAccountKey := "./serviceAccountKey.json"

	exists, err := storage.FileExists(serviceAccountKey)
	if err != nil {
		log.Fatalf("FileExists serviceAccountKey.json check error - %v", err)
	}

	if !exists {
		cmd := exec.Command("gpg", fmt.Sprintf("%s.gpg", serviceAccountKey))
		err = cmd.Run()
		if err != nil {
			log.Fatalf("FileExists serviceAccountKey.json.gpg decryption err - %v", err)
		}
	}

	fstoreClient, err := fstore.NewFirestoreClient(context.Background(), serviceAccountKey)
	if err != nil {
		log.Fatalln(err)
	}
	storageClient := storage.NewFirebaseStorageClient("moda-archive.appspot.com", serviceAccountKey, "./media")
	viewer := viewer.NewViewer(fstoreClient, storageClient)
	plaqueAPIHandler := api.NewPlaqueAPIHandler(viewer)
	go func() {
		log.Fatal(viewer.Startup())
	}()

	log.Fatal(http.ListenAndServe(":8080", plaqueAPIHandler))
}
