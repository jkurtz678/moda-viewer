package viewer

import (
	"log"

	vlcctrl "github.com/CedArctic/go-vlc-ctrl"
)

type VideoPlayer interface {
	initPlayer() error
	playFiles(filepaths []string) error
}

type VLCPlayer struct {
	VLC vlcctrl.VLC
}

func (v *VLCPlayer) initPlayer() error {
	log.Println("VLCPlayer.initPlayer() - running player")
	/* cmd := exec.Command("vlc", "--loop", "--extraintf=http", "--http-port=9090", "--http-password=m0da", "--no-video-title")
	log.Fatalf("VLCPlayer.initPlayer() - error %v", cmd.Run()) */

	vlc, err := vlcctrl.NewVLC("127.0.0.1", 9090, "m0da")
	if err != nil {
		return err
	}
	v.VLC = vlc
	return nil
}

func (v *VLCPlayer) playFiles(filepaths []string) error {
	err := v.VLC.EmptyPlaylist()
	if err != nil {
		return err
	}

	for _, filepath := range filepaths {
		err = v.VLC.Add(filepath)
		if err != nil {
			return err
		}
	}
	return v.VLC.Play(1)
}
