package main

import (
	"context"
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
)

func main() {
	log.Println("MAIN")
	serviceAccountKey := "./serviceAccountKey.json"
	fstoreClient, err := fstore.NewFirestoreClient(context.Background(), serviceAccountKey)
	if err != nil {
		log.Fatalln(err)
	}
	storageClient := storage.NewFirebaseStorageClient("moda-archive.appspot.com", serviceAccountKey, "./media")
	viewer := viewer.NewViewer(fstoreClient, storageClient)
	plaqueAPIHandler := api.NewPlaqueAPIHandler(viewer)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", plaqueAPIHandler))
	}()

	log.Fatal(viewer.Start())
}
