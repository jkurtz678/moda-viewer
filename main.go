package main

import (
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
)

func main() {
	viewer := viewer.NewViewer()
	plaqueAPIHandler := api.NewPlaqueAPIHandler(viewer)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", plaqueAPIHandler))
	}()

	log.Fatal(viewer.Start())
}
