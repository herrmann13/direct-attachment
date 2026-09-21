package upload

import "testing"

func TestPrepareKeepsExistingExtension(t *testing.T) {
	f := Prepare("foto.png", "image/png", 1234)
	if f.Name != "foto.png" {
		t.Errorf("Name = %q, want foto.png", f.Name)
	}
	if f.ContentType != "image/png" {
		t.Errorf("ContentType = %q, want image/png", f.ContentType)
	}
	if f.Size != 1234 {
		t.Errorf("Size = %d, want 1234", f.Size)
	}
}

func TestPrepareAddsExtension(t *testing.T) {
	f := Prepare("foto", "image/jpeg", 10)
	if f.Name != "foto.jpg" {
		t.Errorf("Name = %q, want foto.jpg", f.Name)
	}
}

func TestPrepareEmptyName(t *testing.T) {
	f := Prepare("", "image/webp", 10)
	if f.Name != "photo.webp" {
		t.Errorf("Name = %q, want photo.webp", f.Name)
	}
}

func TestPrepareUnknownType(t *testing.T) {
	f := Prepare("data", "application/octet-stream", 10)
	if f.Name != "data.bin" {
		t.Errorf("Name = %q, want data.bin", f.Name)
	}
}

func TestPrepareStripsPath(t *testing.T) {
	f := Prepare("/tmp/../../evil.png", "image/png", 10)
	if f.Name != "evil.png" {
		t.Errorf("Name = %q, want evil.png", f.Name)
	}
}
