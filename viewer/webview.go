package viewer

import (
	"fmt"

	"github.com/webview/webview"
)

type PlaqueManager interface {
	initPlaque()
	refreshPlaque(tokenMetaID string)
	showPlaque() error
}

type Webview struct {
	webview.WebView
}

func (wv *Webview) initPlaque() {
	logger.Printf("initPlaque()")
	debug := true
	w := webview.New(debug)
	wv.WebView = w
}

func (wv *Webview) refreshPlaque(tokenMetaID string) {
	logger.Printf("refreshPlaque()")
	url := fmt.Sprintf("http://localhost:8080?token_meta_id=%s", tokenMetaID)
	wv.WebView.Dispatch(func() {
		wv.WebView.Eval(fmt.Sprintf(`
		window.location.href = "%s"	
		`, url))
	})
}

func (wv *Webview) showPlaque() error {
	logger.Printf("showPlaque()")
	w := wv.WebView
	defer w.Destroy()
	w.SetTitle("moda")
	w.SetSize(800, 600, webview.HintNone)
	w.Run()
	logger.Printf("end show plaque")
	return nil
}
