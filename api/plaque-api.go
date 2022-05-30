package api

import (
	"encoding/json"
	"fmt"
	"io"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

//PlaqueAPIHandler handles plaque template requests
type PlaqueAPIHandler struct {
	Viewer         *viewer.Viewer
	PlaqueTemplate string
	*httprouter.Router
}

func NewPlaqueAPIHandler(viewer *viewer.Viewer) *PlaqueAPIHandler {
	h := &PlaqueAPIHandler{
		Viewer:         viewer,
		PlaqueTemplate: "template/plaque.html",
		Router:         httprouter.New(),
	}
	h.Router.GET("/", h.servePlaque)
	h.Router.GET("/api/status", h.getStatus)
	return h
}

func (h *PlaqueAPIHandler) servePlaque(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.ServeFile(w, r, h.PlaqueTemplate)
}

func (h *PlaqueAPIHandler) getStatus(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var PlaqueStatus struct {
		Plaque          *fstore.FirestorePlaque    `json:"plaque"`
		ActiveTokenMeta *fstore.FirestoreTokenMeta `json:"active_token_meta"`
	}

	var err error
	PlaqueStatus.Plaque, err = h.Viewer.ReadLocalPlaqueFile()
	if err != nil {
		log.Printf("PlaqueAPIHandler.getStatus - failed to get plaque data %v", err)
		PlaqueStatus.ActiveTokenMeta = nil
	}

	PlaqueStatus.ActiveTokenMeta, err = h.getVLCMeta()
	if err != nil {
		log.Printf("PlaqueAPIHandler.getStatus - failed to get active token meta %v", err)
		PlaqueStatus.ActiveTokenMeta = nil
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(PlaqueStatus); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
	}
}

func (h *PlaqueAPIHandler) getVLCMeta() (*fstore.FirestoreTokenMeta, error) {
	log.Printf("PlaqueAPIHandler.getVLCMeta")

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
	log.Printf("PlaqueAPIHandler.getVLCMeta - resBody %s", resBody)

	var vlcStatus VLCStatus
	err = json.Unmarshal(resBody, &vlcStatus)
	if err != nil {
		return nil, err
	}
	log.Printf("PlaqueAPIHandler.getVLCMeta - vlcUnmarshal %+v", vlcStatus)

	filename := vlcStatus.Information.Category.Meta.Filename
	mediaID := strings.TrimSuffix(filename, filepath.Ext(filename))

	if mediaID == "" {
		return nil, fmt.Errorf("PlaqueAPIHandler.getVLCMeta - empty media id")
	}

	meta, err := h.Viewer.GetTokenMetaForMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

type VLCStatus struct {
	Information struct {
		Category struct {
			Meta struct {
				Filename string `json:"filename"`
			} `json:"meta"`
		} `json:"category"`
	} `json:"information"`
}
