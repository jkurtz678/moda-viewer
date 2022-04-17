package main

import (
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
	"text/template"
)

func main() {
	// init plaque fileserver
	http.HandleFunc("/", servePlaque)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", nil))
	}()

	viewer := viewer.NewViewer()
	log.Fatal(viewer.Start())
}

func servePlaque(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/plaque.html"))

	type PlaqueData struct {
		Title string
		Body  string
	}

	data := PlaqueData{
		Title: "Plaque",
		Body:  "Some body",
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}
