package handler

import (
	"bytes"

	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/config"
)

var (
	pathRootFallback  = []byte("/")
	defaultIndexNames = []string{"index.html"}
)

// 🧊 NewStatic — холодный кеш: всё уже готово, просто отдай и не мешайся.
func NewStatic(from, rootPath string, cfg config.Static) fasthttp.RequestHandler {
	if cfg.Root == "" {
		cfg.Root = rootPath
	}
	if len(cfg.IndexNames) == 0 {
		cfg.IndexNames = defaultIndexNames
	}
	rp := []byte(cfg.Root)
	prefixLen := len(from)
	fs := &fasthttp.FS{
		Root:               "",
		Compress:           cfg.Compress,
		CompressBrotli:     cfg.CompressBrotli,
		CompressZstd:       cfg.CompressZstd,
		GenerateIndexPages: cfg.GenerateIndexPages,
		IndexNames:         cfg.IndexNames,
		CacheDuration:      cfg.CacheDuration,
		AllowEmptyRoot:     true,
		AcceptByteRange:    cfg.AcceptByteRange,
		SkipCache:          cfg.SkipCache,
		PathRewrite: func(ctx *fasthttp.RequestCtx) []byte {
			path := ctx.Path()
			if len(path) >= prefixLen {
				return append(rp, path[prefixLen:]...)
			}
			// fallback на корень
			return pathRootFallback
		},
	}
	h := fs.NewRequestHandler()
	fs.PathNotFound = func(ctx *fasthttp.RequestCtx) {
		if bytes.Equal(ctx.URI().Path(), pathRootFallback) {
			// Защита от бесконечного цикла
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		// fallback для SPA на корень "/" — IndexNames сработают автоматически
		ctx.Request.SetRequestURIBytes(pathRootFallback)
		h(ctx)
	}
	return h
}
