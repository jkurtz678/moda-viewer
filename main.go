package main

import (
	"context"
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
)

func main() {
	fstoreClient, err := fstore.NewFirestoreClient(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	viewer := viewer.NewViewer(fstoreClient)
	plaqueAPIHandler := api.NewPlaqueAPIHandler(viewer)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", plaqueAPIHandler))
	}()

	log.Fatal(viewer.Start())
}
