package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"jkurtz678/moda-viewer/fstore"
	"jkurtz678/moda-viewer/viewer"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/franela/goblin"
)

func TestPlaqueAPI(t *testing.T) {
	g := goblin.Goblin(t)

	tmpdir := t.TempDir()

	// setup test config file
	testConfig := fstore.FirestorePlaque{DocumentID: "1", Plaque: fstore.Plaque{Name: "test"}}
	file, err := json.Marshal(testConfig)
	g.Assert(err).IsNil()
	configPath := filepath.Join(tmpdir, "config_test.json")
	g.Assert(ioutil.WriteFile(configPath, file, 0644)).IsNil()

	// setup test meta file
	testMeta := &fstore.FirestoreTokenMeta{DocumentID: "1", TokenMeta: fstore.TokenMeta{
		Name:   "starry night",
		Artist: "van gogh",
	}}
	file, err = json.Marshal(testMeta)
	g.Assert(err).IsNil()
	metaPath := filepath.Join(tmpdir, "1.json")
	g.Assert(ioutil.WriteFile(metaPath, file, 0644)).IsNil()

	fstoreClientStub := &fstore.FstoreClientStub{}
	v := viewer.NewViewer(fstoreClientStub, nil)
	v.PlaqueFile = configPath
	v.MetadataDir = tmpdir

	h := NewPlaqueAPIHandler(v)
	h.PlaqueTemplate = "../template/plaque.html"

	g.Describe("PlaqueAPIHandler", func() {
		g.It("Should load go template with token meta", func() {
			w, r := testWR("GET", "/", "")
			h.ServeHTTP(w, r)
			g.Assert(w.Code).Equal(200)
			g.Assert(strings.Contains(w.Body.String(), testMeta.TokenMeta.Name)).IsTrue()
			g.Assert(strings.Contains(w.Body.String(), testMeta.TokenMeta.Artist)).IsTrue()
		})
	})
}

func testWR(reqMethod, reqPath, bodyStr string) (*httptest.ResponseRecorder, *http.Request) {
	body := bytes.NewReader([]byte(bodyStr))
	r, err := http.NewRequest(reqMethod, reqPath, body)
	if err != nil {
		panic(err)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	r = r.WithContext(context.Background())

	w := httptest.NewRecorder()

	return w, r
}
