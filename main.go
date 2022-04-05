package main

import (
	"log"
	"net/http"
	"os/exec"
	"text/template"

	"github.com/webview/webview"
)

func main() {
	// init vlc
	cmd := exec.Command("vlc", "playlist.m3u", "--fullscreen", "--loop", "--no-video-title", "--no-macosx-fspanel")
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// init plaque fileserver
	http.HandleFunc("/", servePlaque)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", nil))
	}()

	// init webview
	log.Println("Opening webview....")
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Minimal webview example")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("http://localhost:8080")
	w.Run()
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

type TokenMeta struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

type FirestoreTokenMeta struct {
	DocumentID string    `json:"document_id"`
	TokenMeta  TokenMeta `json:"token_meta"`
}
