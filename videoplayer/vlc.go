package videoplayer

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	vlcctrl "github.com/CedArctic/go-vlc-ctrl"
)

type VideoPlayer interface {
	InitPlayer()
	PlayFiles(filepaths []string) error
	GetStatus() (*VLCStatus, error)
}

type VLCPlayer struct {
	VLC    vlcctrl.VLC
	Client *http.Client
}

func NewVLCPlayer() *VLCPlayer {
	vlc, err := vlcctrl.NewVLC("127.0.0.1", 9090, "m0da")
	// for some reason vlcctrl returns an error that will never fail here, so its safe to log.Fatal
	if err != nil {
		log.Fatal(err)
	}
	return &VLCPlayer{VLC: vlc, Client: &http.Client{Timeout: 5 * time.Second}}
}

func (v *VLCPlayer) InitPlayer() {
	log.Println("VLCPlayer.InitPlayer() - running player")
	cmd := exec.Command("vlc", "--loop", "--extraintf=http", "--http-port=9090", "--http-password=m0da", "--no-video-title")
	log.Fatalf("VLCPlayer.InitPlayer() - error %v", cmd.Run())
}

func (v *VLCPlayer) PlayFiles(filepaths []string) error {
	log.Printf("VLCPlayer.PlayFiles() - playing playlist of %v file(s)", len(filepaths))
	err := v.VLC.EmptyPlaylist()
	if err != nil {
		// empty playlist will typically fail here because the VLC instance has not yet started up, wait a moment and then try again
		log.Printf("VLCPlayer.PlayFiles error: %v, waiting 1 second and then trying again...", err)
		time.Sleep(time.Second)
		return v.PlayFiles(filepaths)
	}

	for _, filepath := range filepaths {
		err = v.VLC.Add(filepath)
		if err != nil {
			return err
		}
	}
	return v.VLC.Play(1)
}

type VLCStatus struct {
	Information Information `json:"information"`
}

type Information struct {
	Category Category `json:"category"`
}
type Category struct {
	Meta Meta `json:"meta"`
}

type Meta struct {
	Filename string `json:"filename"`
}

// GetStatus returns status of vlc instance, such as actively playing file
func (v *VLCPlayer) GetStatus() (*VLCStatus, error) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9090/requests/status.json", http.NoBody)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("", "m0da")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var vlcStatus VLCStatus
	err = json.Unmarshal(resBody, &vlcStatus)
	if err != nil {
		return nil, err
	}

	return &vlcStatus, nil

}
