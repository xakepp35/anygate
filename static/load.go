package static

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/valyala/fasthttp"
)

func LoadFile(fullPath string) (*File, int, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fasthttp.StatusNotFound, err // 404
		case os.IsPermission(err):
			return nil, fasthttp.StatusForbidden, err // 403
		default:
			return nil, fasthttp.StatusInternalServerError, err // 500
		}
	}
	defer f.Close()
	// Отсекаем директории.
	if fi, err := f.Stat(); err != nil || fi.IsDir() {
		return nil, fasthttp.StatusNotFound, err
	}
	payload, err := io.ReadAll(f)
	if err != nil {
		return nil, fasthttp.StatusInternalServerError, err
	}
	contentType := DetectContentType(fullPath, payload)
	return &File{
		Body:        payload,
		ContentType: contentType,
	}, fasthttp.StatusOK, nil
}

func DetectContentType(fullPath string, payload []byte) string {
	ext := path.Ext(fullPath)
	contentType := mime.TypeByExtension(ext)
	if contentType != "" {
		return contentType
	}
	return http.DetectContentType(payload)
}
