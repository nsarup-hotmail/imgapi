package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/nsarup/imgapi/internal/processing"
	"github.com/nsarup/imgapi/pkg/api"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

// handleImages handles POST /images for uploads.
func (s *Server) handleImages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleUpload(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	var (
		data     []byte
		filename string
		err      error
	)

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(s.cfg.MaxUploadBytes); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("missing file field"))
			return
		}
		defer file.Close()
		data, err = processing.CopyLimit(file, s.cfg.MaxUploadBytes)
		if err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, err)
			return
		}
		if header != nil {
			filename = header.Filename
		}
	} else {
		data, err = processing.CopyLimit(r.Body, s.cfg.MaxUploadBytes)
		if err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, err)
			return
		}
		filename = r.Header.Get("X-Filename")
	}

	id, err := s.svc.SaveImage(data, filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, api.UploadResponse{ID: id})
}

// handleGetImage handles GET /images/{id}[.{ext}] with optional Accept negotiation.
func (s *Server) handleGetImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// path after /images/
	tail := strings.TrimPrefix(r.URL.Path, "/images/")
	if tail == "" || tail == "/" {
		http.NotFound(w, r)
		return
	}
	id, ext := splitIDExt(tail)
	target := ""
	if ext != "" {
		switch strings.ToLower(ext) {
		case "jpg", "jpeg":
			target = string(processing.FormatJPEG)
		case "png":
			target = string(processing.FormatPNG)
		default:
			http.Error(w, "unsupported format", http.StatusBadRequest)
			return
		}
	} else {
		// Accept negotiation only when no extension supplied
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "image/jpeg") || strings.Contains(accept, "image/jpg") {
			target = string(processing.FormatJPEG)
		} else if strings.Contains(accept, "image/png") {
			target = string(processing.FormatPNG)
		}
	}

	// parse processing options
	q := r.URL.Query()
	opts := processing.Options{}
	if target != "" {
		opts.Target = processing.SupportedFormat(target)
	}
	if v := q.Get("quality"); v != "" {
		opts.Quality = processing.ParseInt(v)
	}
	if processing.ParseBool(q.Get("gray")) || processing.ParseBool(q.Get("grayscale")) {
		opts.Grayscale = true
	}
	if v := q.Get("w"); v != "" {
		opts.Width = processing.ParseInt(v)
	}
	if v := q.Get("h"); v != "" {
		opts.Height = processing.ParseInt(v)
	}
	if processing.ParseBool(q.Get("thumb")) || processing.ParseBool(q.Get("thumbnail")) {
		opts.Thumbnail = true
	}

	b, ct, err := s.svc.GetImageWithOptions(id, opts)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "not exist") {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func splitIDExt(p string) (string, string) {
	base := path.Base(p)
	dot := strings.LastIndexByte(base, '.')
	if dot <= 0 {
		return base, ""
	}
	return base[:dot], base[dot+1:]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}
