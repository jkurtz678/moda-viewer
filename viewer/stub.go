package viewer

import (
	"jkurtz678/moda-viewer/fstore"
	"sync"
)

type VideoPlayerStub struct {
	filepathPlayed string
	wg             sync.WaitGroup
}

func (v *VideoPlayerStub) playMedia(filepath string, videoStartCallback func(mediaID string)) error {
	v.filepathPlayed = filepath
	v.wg.Done()
	return nil
}

type PlaqueManagerStub struct {
	tokenDisplayed *fstore.FirestoreTokenMeta
}

func (p *PlaqueManagerStub) initPlaque() {
}

func (p *PlaqueManagerStub) showPlaque() error {
	return nil
}

func (p *PlaqueManagerStub) navigateURL(tokenMetaID string) {
}
