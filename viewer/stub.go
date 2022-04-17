package viewer

import (
	"jkurtz678/moda-viewer/fstore"
	"sync"
)

type VideoPlayerStub struct {
	filepathPlayed string
	wg             sync.WaitGroup
}

func (v *VideoPlayerStub) playMedia(filepath string) error {
	v.filepathPlayed = filepath
	v.wg.Done()
	return nil
}

type PlaqueManagerStub struct {
	tokenDisplayed *fstore.FirestoreTokenMeta
}

func (p *PlaqueManagerStub) showPlaque(tokenMeta *fstore.FirestoreTokenMeta) error {
	p.tokenDisplayed = tokenMeta
	return nil
}
