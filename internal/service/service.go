package service

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/nsarup/imgapi/internal/processing"
	"github.com/nsarup/imgapi/internal/storage"
)

// Service wires storage and processing to deliver API behaviors.
type Service struct {
	store storage.Store
}

func New(store storage.Store) *Service {
	return &Service{store: store}
}

// SaveImage persists the provided bytes and returns an image ID.
func (s *Service) SaveImage(data []byte, originalName string) (string, error) {
	ext := filepath.Ext(originalName)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	return s.store.Save(bytesReader(data), ext)
}

// GetImage returns the image bytes, optionally transcoded to target format.
// target can be "", "jpeg", "png".
func (s *Service) GetImage(id string, target string) ([]byte, string, error) {
	b, err := s.store.Load(id)
	if err != nil {
		return nil, "", err
	}
	if target == "" {
		// detect current format to return appropriate Content-Type
		f, derr := processing.DetectFormat(b)
		if derr != nil {
			return b, "application/octet-stream", nil
		}
		switch f {
		case processing.FormatJPEG:
			return b, "image/jpeg", nil
		case processing.FormatPNG:
			return b, "image/png", nil
		default:
			return b, "application/octet-stream", nil
		}
	}
	switch target {
	case string(processing.FormatJPEG):
		out, ct, err := processing.Transcode(b, processing.FormatJPEG)
		return out, ct, err
	case string(processing.FormatPNG):
		out, ct, err := processing.Transcode(b, processing.FormatPNG)
		return out, ct, err
	default:
		return nil, "", errors.New("unsupported target format")
	}
}

// GetImageWithOptions returns the image bytes after applying processing options.
func (s *Service) GetImageWithOptions(id string, opts processing.Options) ([]byte, string, error) {
	b, err := s.store.Load(id)
	if err != nil {
		return nil, "", err
	}
	if opts.IsNoop() {
		// same behavior as GetImage with empty target
		f, derr := processing.DetectFormat(b)
		if derr != nil {
			return b, "application/octet-stream", nil
		}
		switch f {
		case processing.FormatJPEG:
			return b, "image/jpeg", nil
		case processing.FormatPNG:
			return b, "image/png", nil
		default:
			return b, "application/octet-stream", nil
		}
	}
	out, ct, err := processing.Process(b, opts)
	return out, ct, err
}

// bytesReader returns a new reader for the byte slice without escaping the data.
func bytesReader(b []byte) *bytesReaderT { return &bytesReaderT{b: b} }

type bytesReaderT struct {
	b []byte
	i int
}

func (r *bytesReaderT) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
