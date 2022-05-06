package viewer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/nxadm/tail"
)

type VideoPlayer interface {
	playMedia(filepath string, videoStartCallback func(mediaID string)) error
}

type VLCPlayer struct{}

func (v *VLCPlayer) playMedia(filepath string, videoStartCallback func(mediaID string)) error {
	logger.Printf("playMedia() - %s", filepath)
	//cmd := exec.Command("vlc", filepath, "--fullscreen", "--loop", "--no-video-title", "--no-macosx-fspanel")
	go func() {
		err := os.Truncate("vlc.log", 0)
		if err != nil {
			log.Fatal(err)
		}
		t, err := tail.TailFile("vlc.log", tail.Config{Follow: true})
		if err != nil {
			log.Fatal(err)
		}
		for line := range t.Lines {
			fmt.Println("output", line.Text)
			if strings.Contains(line.Text, "successfully opened") && strings.Contains(line.Text, "moda-viewer/media") {
				fmt.Println("found line", line.Text)
				splitSlash := strings.Split(line.Text, "/")
				splitQuote := strings.Split(splitSlash[len(splitSlash)-1], "'")
				mediaID := strings.Split(splitQuote[0], ".")[0]
				videoStartCallback(mediaID)

				// truncate log file after new media is played to prevent it from getting too large
				err := os.Truncate("vlc.log", 0)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
	//cmd := exec.Command("vlc", filepath, "--loop", "--no-video-title", "--no-macosx-fspanel", "--file-logging", "--logfile=vlc.txt", "--log-verbose=3")
	//cmd := exec.Command("vlc", filepath, "--loop", "--intf=http", "--http-port=9090", "--http-password=m0da", "--no-audio", "--no-video-title", "--verbose=2", "--file-logging", "--logfile=vlc.log", "--log-verbose=3", " --http-port 9090")
	cmd := exec.Command("vlc", filepath, "--loop", "--intf=http", "--http-port=9090", "--http-password=m0da", "--no-video-title")
	return cmd.Run()
}
