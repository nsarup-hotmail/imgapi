package httpapi_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nsarup/imgapi/internal/config"
	"github.com/nsarup/imgapi/internal/httpapi"
	"github.com/nsarup/imgapi/internal/logging"
	"github.com/nsarup/imgapi/internal/service"
	"github.com/nsarup/imgapi/internal/storage"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.LoadFromEnv()
	cfg.DataDir = t.TempDir()
	cfg.MaxUploadBytes = 5 * 1024 * 1024
	log := logging.New(io.Discard)
	store, err := storage.NewFileStore(cfg.DataDir)
	if err != nil {
		t.Fatalf("storage: %v", err)
	}
	svc := service.New(store)
	return httpapi.NewServer(cfg, log, svc).Handler()
}

func makePNG(t *testing.T, w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

type uploadResp struct{ ID string `json:"id"` }

func TestUploadAndGetOriginal(t *testing.T) {
	h := newTestServer(t)
	pngBytes := makePNG(t, 4, 4)
	r := httptest.NewRequest(http.MethodPost, "/images", bytes.NewReader(pngBytes))
	r.Header.Set("Content-Type", "application/octet-stream")
	r.Header.Set("X-Filename", "sample.png")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", w.Code, w.Body.String())
	}
	var ur uploadResp
	if err := json.Unmarshal(w.Body.Bytes(), &ur); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if ur.ID == "" { t.Fatal("missing id") }

	// Retrieve original (no extension, no Accept)
	gr := httptest.NewRequest(http.MethodGet, "/images/"+ur.ID, nil)
	gw := httptest.NewRecorder()
	h.ServeHTTP(gw, gr)
	if gw.Code != http.StatusOK {
		t.Fatalf("get status=%d", gw.Code)
	}
	if ct := gw.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestGetWithExtensionJPEG(t *testing.T) {
	h := newTestServer(t)
	pngBytes := makePNG(t, 2, 2)
	u := httptest.NewRequest(http.MethodPost, "/images", bytes.NewReader(pngBytes))
	u.Header.Set("Content-Type", "application/octet-stream")
	u.Header.Set("X-Filename", "x.png")
	uw := httptest.NewRecorder()
	h.ServeHTTP(uw, u)
	var ur uploadResp
	_ = json.Unmarshal(uw.Body.Bytes(), &ur)

	gr := httptest.NewRequest(http.MethodGet, "/images/"+ur.ID+".jpg", nil)
	gw := httptest.NewRecorder()
	h.ServeHTTP(gw, gr)
	if gw.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", gw.Code, gw.Body.String())
	}
	if ct := gw.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestGetWithAcceptHeader(t *testing.T) {
	h := newTestServer(t)
	pngBytes := makePNG(t, 2, 2)
	u := httptest.NewRequest(http.MethodPost, "/images", bytes.NewReader(pngBytes))
	u.Header.Set("Content-Type", "application/octet-stream")
	u.Header.Set("X-Filename", "x.png")
	uw := httptest.NewRecorder()
	h.ServeHTTP(uw, u)
	var ur uploadResp
	_ = json.Unmarshal(uw.Body.Bytes(), &ur)

	gr := httptest.NewRequest(http.MethodGet, "/images/"+ur.ID, nil)
	gr.Header.Set("Accept", "image/jpeg")
	gw := httptest.NewRecorder()
	h.ServeHTTP(gw, gr)
	if gw.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", gw.Code, gw.Body.String())
	}
	if ct := gw.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}
