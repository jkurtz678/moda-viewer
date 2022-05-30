package videoplayer

import (
	"io/ioutil"
	"os"
	"strings"
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
	v.ActivePlaylistFilepaths = filepaths
	v.PlayFilesWaitGroup.Done()
	return nil
}

func parsePlaylistFile(filepath string) ([]string, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(byteValue), "\n")
	return files, nil
}
