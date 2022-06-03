package videoplayer

import (
	"log"
	"os/exec"

	vlcctrl "github.com/CedArctic/go-vlc-ctrl"
)

type VideoPlayer interface {
	InitPlayer()
	PlayFiles(filepaths []string) error
}

type VLCPlayer struct {
	VLC vlcctrl.VLC
}

func (v *VLCPlayer) InitPlayer() {
	log.Println("VLCPlayer.InitPlayer() - running player")

	vlc, err := vlcctrl.NewVLC("127.0.0.1", 9090, "m0da")
	if err != nil {
		log.Fatal(err)
	}
	v.VLC = vlc

	cmd := exec.Command("vlc", "--loop", "--extraintf=http", "--http-port=9090", "--http-password=m0da", "--no-video-title")
	log.Fatalf("VLCPlayer.InitPlayer() - error %v", cmd.Run())
}

func (v *VLCPlayer) PlayFiles(filepaths []string) error {
	log.Printf("VLCPlayer.PlayFiles() - playing playlist of %v files", len(filepaths))
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
