package viewer

import (
	"jkurtz678/moda-viewer/fstore"

	"github.com/webview/webview"
)

type PlaqueManager interface {
	showPlaque(meta *fstore.FirestoreTokenMeta) error
}

type Webview struct{}

func (wv *Webview) showPlaque(meta *fstore.FirestoreTokenMeta) error {
	logger.Printf("showPlaque() - %s", meta.TokenMeta.Name)
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Minimal webview example")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("http://localhost:8080")
	w.Run()
	logger.Printf("end show plaque")
	return nil
}
