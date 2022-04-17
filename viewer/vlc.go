package viewer

import (
	"os/exec"
)

type VideoPlayer interface {
	playMedia(filepath string) error
}

type VLCPlayer struct{}

func (v *VLCPlayer) playMedia(filepath string) error {
	logger.Printf("playMedia() - %s", filepath)
	cmd := exec.Command("vlc", filepath, "--fullscreen", "--loop", "--no-video-title", "--no-macosx-fspanel")
	return cmd.Run()
}
