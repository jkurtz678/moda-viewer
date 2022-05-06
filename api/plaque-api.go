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
		metaID = h.getVLCMetaID()
	}
	meta, err := h.Viewer.ReadMetadata(metaID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("executing template for token meta %+v", meta)
	err = tmpl.Execute(w, meta.TokenMeta)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *PlaqueAPIHandler) getActiveToken(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type ActiveTokenResp struct {
		ActiveTokenMetaID string `json:"active_token_meta_id"`
	}

	activeTokenID := h.getVLCMetaID()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ActiveTokenResp{ActiveTokenMetaID: activeTokenID}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
	}
}

func (h *PlaqueAPIHandler) getVLCMetaID() string {
	client := http.Client{Timeout: 5 * time.Second}

    req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9090/requests/status.xml", http.NoBody)
    if err != nil {
        log.Fatal(err)
    }

	req.SetBasicAuth("", "m0da")

    res, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }

    defer res.Body.Close()

    resBody, err := io.ReadAll(res.Body)
    if err != nil {
        log.Fatal(err)
    }

	firstParse := strings.Split(string(resBody), "<info name='filename'>media\\")[1]
	secondParse := strings.Split(firstParse, "</info>")[0]
	mediaID := strings.Split(secondParse, ".")[0]

	meta, err := h.Viewer.GetTokenMetaForMediaID(mediaID)
	if err != nil {
		log.Fatal(err)
	}	
	return meta.DocumentID
}