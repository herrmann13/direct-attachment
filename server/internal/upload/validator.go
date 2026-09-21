// Package upload prepares the opaque (end-to-end encrypted) payload before it
// is relayed to the PC. Because the file is encrypted, the server must not
// inspect its contents; this package only normalizes the file name and carries
// the client-declared metadata.
package upload

import (
	"path/filepath"
	"strings"
)

// File is the metadata the server relays alongside the encrypted bytes.
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// Prepare normalizes the destination file name and returns the relay metadata.
// It performs no content inspection; type/size enforcement happens on the
// phone (before encryption) and on the PC (after decryption).
func Prepare(name, contentType string, size int64) File {
	return File{
		Name:        normalizeName(name, contentType),
		ContentType: contentType,
		Size:        size,
	}
}

// normalizeName strips directory components and guarantees a sensible
// extension based on the declared content type, falling back to "photo".
func normalizeName(name, contentType string) string {
	name = filepath.Base(strings.TrimSpace(name))
	ext := extensionFor(contentType)

	if name == "" || name == "." || name == "/" || name == "\\" {
		return "photo" + ext
	}
	if filepath.Ext(name) == "" {
		name += ext
	}
	return name
}

// extensionFor maps a content type to a canonical extension, with a safe
// default for unknown/opaque types.
func extensionFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
