package webview

import (
	"log"
	"os/exec"
)

type PlaqueManager interface {
	InitPlaque()
	//navigateURL(tokenMetaID string)
}

type PythonWebview struct {
}

func (pw *PythonWebview) InitPlaque() {
	log.Printf("PythonWebview.InitPlaque() - running plaque webview")
	cmd := exec.Command("python3", "viewer/plaque_webview.py", "http://localhost:8080")
	log.Fatalf("PythonWebview.InitPlaque() - error %v", cmd.Run())
}

/* func (pq *PythonWebview) navigateURL(tokenMetaID string) {
	log.Printf("PythonWebview.navigateURL - %s", tokenMetaID)
} */
