package upload

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/direct-attachment-plugin/server/internal/httpx"
)

// transparentPNG is a valid 1x1 PNG image.
const transparentPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func mustDecode(t *testing.T, b64 string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	return b
}

func TestValidateImagePNG(t *testing.T) {
	data := mustDecode(t, transparentPNG)
	f, err := ValidateImage("foto.png", data)
	if err != nil {
		t.Fatalf("ValidateImage() error = %v", err)
	}
	if f.ContentType != "image/png" {
		t.Errorf("ContentType = %q, want image/png", f.ContentType)
	}
	if f.Name != "foto.png" {
		t.Errorf("Name = %q, want foto.png", f.Name)
	}
	if f.Size != int64(len(data)) {
		t.Errorf("Size = %d, want %d", f.Size, len(data))
	}
}

func TestValidateImageAddsExtension(t *testing.T) {
	data := mustDecode(t, transparentPNG)
	f, err := ValidateImage("foto", data)
	if err != nil {
		t.Fatalf("ValidateImage() error = %v", err)
	}
	if f.Name != "foto.png" {
		t.Errorf("Name = %q, want foto.png", f.Name)
	}
}

func TestValidateImageEmptyName(t *testing.T) {
	data := mustDecode(t, transparentPNG)
	f, err := ValidateImage("", data)
	if err != nil {
		t.Fatalf("ValidateImage() error = %v", err)
	}
	if f.Name != "photo.png" {
		t.Errorf("Name = %q, want photo.png", f.Name)
	}
}

func TestValidateImageRejectsNonImage(t *testing.T) {
	if _, err := ValidateImage("note.txt", []byte("hello, this is text")); err == nil {
		t.Fatal("expected error for non-image data")
	} else {
		var apiErr *httpx.Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("error type = %T, want *httpx.Error", err)
		}
		if apiErr.Code != "unsupported_type" {
			t.Errorf("code = %q, want unsupported_type", apiErr.Code)
		}
	}
}

func TestValidateImageRejectsEmpty(t *testing.T) {
	if _, err := ValidateImage("x.png", []byte{}); err == nil {
		t.Fatal("expected error for empty data")
	}
}
