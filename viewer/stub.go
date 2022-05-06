package viewer

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type VideoPlayerStub struct {
	filepathPlayed string
	wg             sync.WaitGroup
}

func (v *VideoPlayerStub) playMedia(filepath string, videoStartCallback func(mediaID string)) error {
	v.filepathPlayed = filepath

	files, err := parsePlaylistFile(filepath)
	if err != nil {
		return err
	}
	slashSplit := strings.Split(files[0], "/")
	mediaID := strings.Split(slashSplit[len(slashSplit)-1], ".")[0]

	videoStartCallback(mediaID)

	v.wg.Done()
	return nil
}

func parsePlaylistFile(filepath string) ([]string, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(byteValue), "\n")
	return files, nil
}

type PlaqueManagerStub struct {
	tokenMetaIDNavigated string
}

func (p *PlaqueManagerStub) initPlaque() {
}

func (p *PlaqueManagerStub) showPlaque() error {
	return nil
}

func (p *PlaqueManagerStub) navigateURL(tokenMetaID string) {
	p.tokenMetaIDNavigated = tokenMetaID
}
