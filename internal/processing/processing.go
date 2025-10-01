package processing

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

// SupportedFormat represents a canonical image format name.
type SupportedFormat string

const (
	FormatJPEG SupportedFormat = "jpeg"
	FormatPNG  SupportedFormat = "png"
)

var errUnsupported = errors.New("unsupported format")

// DetectFormat tries to detect the image format from bytes using stdlib image.Registered formats.
func DetectFormat(b []byte) (SupportedFormat, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return FormatJPEG, nil
	case "png":
		return FormatPNG, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupported, format)
	}
}

// Transcode converts image bytes to the requested target format.
func Transcode(in []byte, target SupportedFormat) ([]byte, string, error) {
	img, _, err := image.Decode(bytes.NewReader(in))
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	switch target {
	case FormatJPEG:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	case FormatPNG:
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil
	default:
		return nil, "", errUnsupported
	}
}

// CopyLimit copies up to n bytes from r to a buffer; returns error if exceeded.
func CopyLimit(r io.Reader, n int64) ([]byte, error) {
	var buf bytes.Buffer
	// LimitReader allows at most n bytes; we detect overflow by reading one extra byte
	limited := io.LimitedReader{R: r, N: n + 1}
	if _, err := io.Copy(&buf, &limited); err != nil {
		return nil, err
	}
	if int64(buf.Len()) > n {
		return nil, fmt.Errorf("payload too large: limit %d bytes", n)
	}
	return buf.Bytes(), nil
}

// Options define processing/transformation parameters.
type Options struct {
	Target    SupportedFormat // "jpeg" or "png"; empty means keep original
	Quality   int             // 1-100 for JPEG; 0 means default 85
	Grayscale bool
	Width     int  // resize/thumbnail width if > 0
	Height    int  // resize/thumbnail height if > 0
	Thumbnail bool // if true and both dims specified, do center-crop thumbnail
}

// IsNoop returns true if the options request no transformation and no target change.
func (o Options) IsNoop() bool {
	return !o.Grayscale && o.Width == 0 && o.Height == 0 && o.Thumbnail == false && o.Target == ""
}

// ParseBool accepts "1", "true", "yes" as true (case-insensitive).
func ParseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

// ParseInt returns integer value or 0 on error.
func ParseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// Process applies transformations and encodes to the requested or original format.
func Process(in []byte, opts Options) ([]byte, string, error) {
	// If no options, return early with detected content type
	if opts.IsNoop() {
		f, err := DetectFormat(in)
		if err != nil {
			return in, "application/octet-stream", nil
		}
		switch f {
		case FormatJPEG:
			return in, "image/jpeg", nil
		case FormatPNG:
			return in, "image/png", nil
		default:
			return in, "application/octet-stream", nil
		}
	}

	img, _, err := image.Decode(bytes.NewReader(in))
	if err != nil {
		return nil, "", err
	}

	// filters
	if opts.Grayscale {
		img = imaging.Grayscale(img)
	}

	// resizing
	if opts.Width > 0 || opts.Height > 0 {
		if opts.Thumbnail && opts.Width > 0 && opts.Height > 0 {
			img = imaging.Thumbnail(img, opts.Width, opts.Height, imaging.Lanczos)
		} else {
			w := opts.Width
			h := opts.Height
			img = imaging.Resize(img, w, h, imaging.Lanczos)
		}
	}

	// choose output format
	target := opts.Target
	if target == "" {
		// keep original
		if f, err := DetectFormat(in); err == nil {
			target = f
		} else {
			target = FormatPNG
		}
	}

	var buf bytes.Buffer
	switch target {
	case FormatJPEG:
		q := opts.Quality
		if q <= 0 || q > 100 {
			q = 85
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	case FormatPNG:
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil
	default:
		return nil, "", errUnsupported
	}
}
