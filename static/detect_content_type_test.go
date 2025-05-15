package static_test

import (
	"testing"

	"github.com/xakepp35/anygate/static"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		payload    []byte
		wantPrefix string
	}{
		// Статика React SPA
		{"JS file", "app.js", []byte("console.log('hello');"), "text/javascript; charset=utf-8"},
		{"CSS file", "style.css", []byte("body { margin: 0; }"), "text/css; charset=utf-8"},
		{"HTML file", "index.html", []byte("<!DOCTYPE html><html></html>"), "text/html; charset=utf-8"},
		{"JSON file", "manifest.json", []byte(`{"name": "app"}`), "application/json"},
		{"SVG image", "icon.svg", []byte(`<svg></svg>`), "image/svg+xml"},

		// Картинки
		{"PNG file", "logo.png", []byte{0x89, 0x50, 0x4E, 0x47}, "image/png"},
		{"JPG file", "photo.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"GIF file", "anim.gif", []byte("GIF89a"), "image/gif"},

		// Fallback для неизвестного расширения
		{"Unknown ext, fallback to detect", "data.unknown", []byte("<html><body></body></html>"), "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := static.DetectContentType("/mock/"+tt.filename, tt.payload)
			if ct != tt.wantPrefix {
				t.Errorf("expected %q, got %q", tt.wantPrefix, ct)
			}
		})
	}
}
