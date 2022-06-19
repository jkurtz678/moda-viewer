package api

import (
	"encoding/json"
	"fmt"
	"jkurtz678/moda-viewer/viewer"
	"net/http"

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
		PlaqueTemplate: "ui/plaque.html",
		Router:         httprouter.New(),
	}
	h.Router.GET("/", h.servePlaque)
	h.Router.GET("/api/status", h.getStatus)
	h.Router.ServeFiles("/ui/*filepath", http.Dir("ui"))
	return h
}

func (h *PlaqueAPIHandler) servePlaque(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	http.ServeFile(w, r, h.PlaqueTemplate)
}

func (h *PlaqueAPIHandler) getStatus(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	stateData := h.Viewer.GetViewerState()
	//log.Printf("PlaqueAPIHandler.getStatus state returned - %s", stateData.State)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stateData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("internal error %s", err))
	}
}
