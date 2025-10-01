package storage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Store defines operations for persisting and retrieving image bytes by ID.
type Store interface {
	Save(reader io.Reader, hintedExt string) (id string, err error)
	Load(id string) (bytes []byte, err error)
	PathFor(id string) (string, error)
}

// FileStore stores image files on the local filesystem.
type FileStore struct {
	baseDir string
}

// NewFileStore creates a new FileStore rooted at baseDir.
func NewFileStore(baseDir string) (*FileStore, error) {
	if baseDir == "" {
		return nil, errors.New("baseDir required")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{baseDir: baseDir}, nil
}

// Save writes the content to a new uniquely named file and returns its ID.
func (s *FileStore) Save(reader io.Reader, hintedExt string) (string, error) {
	id := generateID()
	// store original as .bin; we keep extension information separate via processing
	filename := id + ".bin"
	if hintedExt != "" {
		// purely cosmetic; retrieval is by id, processing detects type
		filename = id + "." + sanitizeExt(hintedExt)
	}
	path := filepath.Join(s.baseDir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return "", err
	}
	return id, nil
}

// Load reads the content for a given id by locating a file with known patterns.
func (s *FileStore) Load(id string) ([]byte, error) {
	// try known patterns
	candidates, err := filepath.Glob(filepath.Join(s.baseDir, id+".*"))
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(candidates[0])
}

// PathFor returns a path to the stored file for id.
func (s *FileStore) PathFor(id string) (string, error) {
	candidates, err := filepath.Glob(filepath.Join(s.baseDir, id+".*"))
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", os.ErrNotExist
	}
	return candidates[0], nil
}

func sanitizeExt(ext string) string {
	out := make([]rune, 0, len(ext))
	for _, r := range ext {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return "bin"
	}
	return string(out)
}

func generateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// extremely unlikely; fallback to timestamp-based randomness via hex of zeroes
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}
