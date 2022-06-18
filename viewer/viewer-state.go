package viewer

import (
	"fmt"
	"jkurtz678/moda-viewer/fstore"
	"path/filepath"
	"strings"
)

// ViewerState is the current state of the viewer, corresponds to a different UI shown on the plaque
type ViewerState string

const (
	ViewerStateLoading       = ViewerState("loading")         // plaque is actively trying to display media, most time here is downloading media or metadata files
	ViewerStateQrScan        = ViewerState("qr_scan")         // plaque has no connected wallet address and is displaying a qr code allowing users to scan and gain control
	ViewerStateNoValidTokens = ViewerState("no_valid_tokens") // plaque is connected to a user but has no assigned tokens
	ViewerStateDisplay       = ViewerState("display")         // plaque is showing art and running as normal
	ViewerStateError         = ViewerState("error")           // plaque has encountered an error, will pause breifly and retry
)

type ViewerStateData struct {
	State           ViewerState                `json:"state"`
	Plaque          *fstore.FirestorePlaque    `json:"plaque"`
	ActiveTokenMeta *fstore.FirestoreTokenMeta `json:"active_token_meta"`
}

// GetViewerState
func (v *Viewer) GetViewerState() *ViewerStateData {

	// if v.loadErr is not empty, return ViewerState
	v.stateLock.Lock()
	loadErr := v.loadErr
	v.stateLock.Unlock()
	if loadErr != nil {
		return &ViewerStateData{State: ViewerStateError}
	}

	// if v.loading is true, return ViewerState
	v.stateLock.Lock()
	loading := v.loading
	v.stateLock.Unlock()
	if loading {
		return &ViewerStateData{State: ViewerStateLoading}
	}

	localPlaque, err := v.ReadLocalPlaqueFile()
	if err != nil {
		logger.Printf("GetViewerState - failed to get plaque data %v", err)
		v.stateLock.Lock()
		v.loadErr = err
		v.stateLock.Unlock()
		return &ViewerStateData{State: ViewerStateError}
	}
	// no wallet address means that plaque is not attached to a user, show qr scan
	// return plaque since its data should be used by qr code scan
	if localPlaque.Plaque.WalletAddress == "" {
		return &ViewerStateData{State: ViewerStateQrScan, Plaque: localPlaque}
	}

	// check if tokens are valid, if none exist show no valid tokens
	validTokens := v.getValidTokens(localPlaque.Plaque.TokenMetaIDList)
	if len(validTokens) == 0 {
		return &ViewerStateData{State: ViewerStateNoValidTokens, Plaque: localPlaque}
	}

	activeToken, err := v.getActivelyPlayingToken()
	if err != nil {
		logger.Printf("GetViewerState - failed to get actively playing token with error: %v", err)
		return &ViewerStateData{State: ViewerStateLoading, Plaque: localPlaque}
	}

	// if no states were found above plaque is properly displaying art
	return &ViewerStateData{State: ViewerStateDisplay, Plaque: localPlaque, ActiveTokenMeta: activeToken}
}

// getActivelyPlayingToken will return actively playing token meta
func (v *Viewer) getActivelyPlayingToken() (*fstore.FirestoreTokenMeta, error) {
	playerStatus, err := v.VideoPlayer.GetStatus()
	if err != nil {
		return nil, err
	}

	filename := playerStatus.Information.Category.Meta.Filename
	mediaID := strings.TrimSuffix(filename, filepath.Ext(filename))

	if mediaID == "" {
		return nil, fmt.Errorf("PlaqueAPIHandler.getVLCMeta - empty media id")
	}

	meta, err := v.GetTokenMetaForMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
