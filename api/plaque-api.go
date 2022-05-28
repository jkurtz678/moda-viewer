package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"jkurtz678/moda-viewer/viewer"
	"log"
	"net/http"
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
	h.Router.GET("/api/active-token", h.getActiveToken)
	return h
}

func (h *PlaqueAPIHandler) servePlaque(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	tmpl := template.Must(template.ParseFiles(h.PlaqueTemplate))

	metaID := r.URL.Query().Get("token_meta_id")
	if metaID == "" {
		log.Printf("PlaqueAPIHandler.servePlaque - no token_media_id provided, loading active")
		vlcMeta, err := h.getVLCMetaID()
		if err != nil {
			log.Printf("plaqueAPIHandler.servePlaque - getVLCMetaID error %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
			return
		}
		log.Printf("VLC META %s", vlcMeta)
		metaID = vlcMeta
	}
	meta, err := h.Viewer.ReadMetadata(metaID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
		return
	}

	log.Printf("executing template for token meta %+v", meta)
	err = tmpl.Execute(w, meta.TokenMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
		return
	}
}

func (h *PlaqueAPIHandler) getActiveToken(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type ActiveTokenResp struct {
		ActiveTokenMetaID string `json:"active_token_meta_id"`
	}

	activeTokenID, err := h.getVLCMetaID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ActiveTokenResp{ActiveTokenMetaID: activeTokenID}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
	}
}

func (h *PlaqueAPIHandler) getVLCMetaID() (string, error) {
	log.Printf("PlaqueAPIHandler.getVLCMetaID")

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9090/requests/status.xml", http.NoBody)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth("", "m0da")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	log.Printf("PlaqueAPIHandler.getVLCMetaID - resBody %s", resBody)

	firstSplit := strings.Split(string(resBody), "<info name='filename'>media\\")
	if len(firstSplit) < 2 {
		return "", fmt.Errorf("PlaqueAPIHandler.getVLCMetaID - failed to parse media filename")
	}
	firstParse := firstSplit[1]
	secondParse := strings.Split(firstParse, "</info>")[0]
	mediaID := strings.Split(secondParse, ".")[0]

	if mediaID == "" {
		return "", fmt.Errorf("PlaqueAPIHandler.getVLCMetaID - empty media id")
	}

	meta, err := h.Viewer.GetTokenMetaForMediaID(mediaID)
	if err != nil {
		return "", err
	}
	return meta.DocumentID, nil
}
