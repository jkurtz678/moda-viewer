package main

import (
	"context"
	"jkurtz678/moda-viewer/api"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/storage"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
	"os/exec"
)

func main() {

	//TODO decrypt gpg file

	// install python dependencies
	log.Printf("Checking python dependencies...")
	cmd := exec.Command("pip", "install", "-r", "webview/requirements.txt")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("pip dependency install error - %v", err)
	}

	// add vlc to path if not found, currently not working
	/* log.Printf("Checking for VLC in path...")
	_, err = exec.LookPath("vlc")
	if err != nil {
		log.Printf("VLC not found in path, attempting to add...")
		//cmd := exec.Command("export", `PATH=$PATH:"/C/Program Files/VideoLAN/VLC/"`)
		cmd := exec.Command("source", "videoplayer/vlc-path.sh")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("main failed to add vlc to path: %s - %v", fmt.Sprint(err), stderr.String())

		}
		log.Printf("VLC successfully added to path")
	} else {
		log.Printf("VLC found in path")
	} */

	serviceAccountKey := "./serviceAccountKey.json"
	fstoreClient, err := fstore.NewFirestoreClient(context.Background(), serviceAccountKey)
	if err != nil {
		log.Fatalln(err)
	}
	storageClient := storage.NewFirebaseStorageClient("moda-archive.appspot.com", serviceAccountKey, "./media")
	viewer := viewer.NewViewer(fstoreClient, storageClient)
	plaqueAPIHandler := api.NewPlaqueAPIHandler(viewer)
	go func() {
		log.Fatal(viewer.Startup())
	}()

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", plaqueAPIHandler))
}
