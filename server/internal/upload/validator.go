// Package upload validates uploaded images before they are relayed to the PC.
// It never trusts the client-declared MIME type; the real type is detected by
// sniffing the file signature.
package upload

import (
	"path/filepath"
	"strings"

	"github.com/direct-attachment-plugin/server/internal/httpx"
)

// allowedTypes maps a detected content type to its canonical file extension.
var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// File is a validated image ready to be relayed.
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// ValidateImage detects the real content type of data, enforces the image
// allowlist, and normalizes the destination file name.
func ValidateImage(name string, data []byte) (File, error) {
	if len(data) == 0 {
		return File{}, httpx.BadRequest("bad_request", "empty file")
	}

	contentType := detectContentType(data)
	ext, ok := allowedTypes[contentType]
	if !ok {
		return File{}, httpx.UnsupportedMediaType("unsupported_type", "only JPEG, PNG and WebP images are allowed")
	}

	return File{
		Name:        normalizeName(name, ext),
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

// normalizeName strips any directory components and guarantees a sensible
// extension, falling back to "photo" when the client sends nothing usable.
func normalizeName(name, ext string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == "/" || name == "\\" {
		return "photo" + ext
	}
	if filepath.Ext(name) == "" {
		name += ext
	}
	return name
}
