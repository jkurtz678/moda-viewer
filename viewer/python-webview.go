package viewer

import (
	"log"
	"os/exec"
)

type PythonWebview struct {
}

func (pw *PythonWebview) initPlaque() {
	log.Printf("PythonWebview.initPlaque() - running plaque webview")
	cmd := exec.Command("python3", "viewer/plaque_webview.py", "http://localhost:8080")
	log.Fatalf("PythonWebview.initPlaque() - error %v", cmd.Run())
}

func (pq *PythonWebview) navigateURL(tokenMetaID string) {
	log.Printf("PythonWebview.navigateURL - %s", tokenMetaID)
}
