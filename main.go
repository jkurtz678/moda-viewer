package main

import (
	"context"
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
	"os/exec"
)

func main() {

	//TODO decrypt gpg file

	// install python dependencies
	cmd := exec.Command("pip", "install", "-r", "webview/requirements.txt")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("pip dependency install error - %v", err)
	}

	serviceAccountKey := "./serviceAccountKey.json"
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
