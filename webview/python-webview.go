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
	// check if python3 command exists
	_, err := exec.LookPath("python3")
	if err == nil {
		// python3 command exists, use that to run python
		cmd := exec.Command("python3", "webview/plaque_webview.py", "http://localhost:8080")
		log.Fatalf("PythonWebview.InitPlaque() - error %v", cmd.Run())
		return
	}

	// python3 does not exist, use python as argument
	cmd := exec.Command("python", "webview/plaque_webview.py", "http://localhost:8080")
	log.Fatalf("PythonWebview.InitPlaque() - error %v", cmd.Run())
}

/* func (pq *PythonWebview) navigateURL(tokenMetaID string) {
	log.Printf("PythonWebview.navigateURL - %s", tokenMetaID)
} */
