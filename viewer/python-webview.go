package viewer

import (
	"log"
	"os/exec"
)

type PythonWebview struct {
}

func (pw *PythonWebview) initPlaque() {
	cmd := exec.Command("python", "viewer/plaque_webview.py", "http://localhost:8080")
	log.Println("PythonWebview.initPlaque - ", cmd.Run())
}

func (pq *PythonWebview) navigateURL(tokenMetaID string) {
	log.Printf("PythonWebview.navigateURL - %s", tokenMetaID)
}