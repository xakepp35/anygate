package utils_test

import (
	"testing"

	"github.com/xakepp35/anygate/utils"
)

func TestPathBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		fromLen  int
		ctxPath  []byte
		expected string
	}{
		{
			name:     "Normal path with prefix",
			rootPath: "/var/www",
			fromLen:  7,
			ctxPath:  []byte("/static/app.js"),
			expected: "/var/www/app.js",
		},
		{
			name:     "Empty suffix after prefix",
			rootPath: "/var/www",
			fromLen:  7,
			ctxPath:  []byte("/static"),
			expected: "/var/www",
		},
		{
			name:     "Path shorter than prefix",
			rootPath: "/var/www",
			fromLen:  10,
			ctxPath:  []byte("/foo"),
			expected: "/var/www",
		},
		{
			name:     "Trailing slash after prefix",
			rootPath: "/var/www",
			fromLen:  7,
			ctxPath:  []byte("/static/"),
			expected: "/var/www/",
		},
		{
			name:     "Nested path after prefix",
			rootPath: "/root",
			fromLen:  7,
			ctxPath:  []byte("/prefix/images/logo.png"),
			expected: "/root/images/logo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := utils.NewPathBuilder(tt.rootPath, tt.fromLen, 64)
			result := builder.Build(tt.ctxPath)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPathBuilder_Parallel(t *testing.T) {
	builder := utils.NewPathBuilder("/base", 7, 64)
	paths := [][]byte{
		[]byte("/static/a.js"),
		[]byte("/static/b.js"),
		[]byte("/static/img/logo.png"),
	}
	expected := []string{
		"/base/a.js",
		"/base/b.js",
		"/base/img/logo.png",
	}

	t.Parallel()
	for i := range paths {
		i := i // захват переменной
		t.Run("parallel-"+expected[i], func(t *testing.T) {
			t.Parallel()
			result := builder.Build(paths[i])
			if result != expected[i] {
				t.Errorf("got %q, want %q", result, expected[i])
			}
		})
	}
}
