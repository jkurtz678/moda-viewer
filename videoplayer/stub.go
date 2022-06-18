package videoplayer

import (
	"log"
	"net/url"
	"path/filepath"
	"sync"
)

type VideoPlayerStub struct {
	PlayerInit              bool
	ActivePlaylistFilepaths []string
	PlayFilesWaitGroup      sync.WaitGroup
}

func (v *VideoPlayerStub) InitPlayer() {
	v.PlayerInit = true
}

func (v *VideoPlayerStub) PlayFiles(filepaths []string) error {
	log.Printf("%+v", v.ActivePlaylistFilepaths)
	v.ActivePlaylistFilepaths = filepaths

	if len(filepaths) == 1 && filepaths[0] == "moda-logo.png" {
		return nil
	}
	v.PlayFilesWaitGroup.Done()
	return nil
}

// return ffirst filename in list, need to decode query string because we encode when sending to vlc
func (v *VideoPlayerStub) GetStatus() (*VLCStatus, error) {
	if len(v.ActivePlaylistFilepaths) > 0 {
		unescape, err := url.QueryUnescape(v.ActivePlaylistFilepaths[0])
		if err != nil {
			return nil, err
		}
		filename := filepath.Base(unescape)
		return &VLCStatus{
			Information: Information{
				Category{
					Meta{
						Filename: filename,
					},
				},
			},
		}, nil
	}
	return nil, nil
}
