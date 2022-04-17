package api

import (
	"html/template"
	"jkurtz678/moda-viewer/viewer"
	"log"
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
		PlaqueTemplate: "template/plaque.html",
		Router:         httprouter.New(),
	}
	h.Router.GET("/", h.servePlaque)
	return h
}

func (h *PlaqueAPIHandler) servePlaque(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	tmpl := template.Must(template.ParseFiles(h.PlaqueTemplate))

	tokenMeta, err := h.Viewer.GetActiveTokenMeta()
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, tokenMeta.TokenMeta)
	if err != nil {
		log.Fatal(err)
	}
}
