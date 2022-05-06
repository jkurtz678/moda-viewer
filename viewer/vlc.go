package viewer

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hpcloud/tail"
)

type VideoPlayer interface {
	playMedia(filepath string, videoStartCallback func(mediaID string)) error
}

type VLCPlayer struct{}

func (v *VLCPlayer) playMedia(filepath string, videoStartCallback func(mediaID string)) error {
	logger.Printf("playMedia() - %s", filepath)
	//cmd := exec.Command("vlc", filepath, "--fullscreen", "--loop", "--no-video-title", "--no-macosx-fspanel")
	go func() {
		err := os.Truncate("vlc.txt", 0)
		if err != nil {
			log.Fatal(err)
		}
		t, err := tail.TailFile("vlc.txt", tail.Config{Follow: true})
		if err != nil {
			log.Fatal(err)
		}
		for line := range t.Lines {
			if strings.Contains(line.Text, "successfully opened") && strings.Contains(line.Text, "moda-viewer/media") {
				//fmt.Println(line.Text)
				splitSlash := strings.Split(line.Text, "/")
				splitQuote := strings.Split(splitSlash[len(splitSlash)-1], "'")
				mediaID := strings.Split(splitQuote[0], ".")[0]
				videoStartCallback(mediaID)

				// truncate log file after new media is played to prevent it from getting too large
				err := os.Truncate("vlc.txt", 0)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
	//cmd := exec.Command("vlc", filepath, "--loop", "--no-video-title", "--no-macosx-fspanel", "--file-logging", "--logfile=vlc.txt", "--log-verbose=3")
	cmd := exec.Command("vlc", filepath, "--loop", "--no-video-title", "--file-logging", "--logfile=vlc.txt", "--log-verbose=3")
	return cmd.Run()
}
