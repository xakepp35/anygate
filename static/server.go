package static

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/utils"
)

type Config struct {
	RootPath     string // "./dist/"
	FromLen      int    // len("/static/")
	CapHint      int    // средняя длина пути
	DisableCache bool   // true = без кеша
	FallbackName string // "index.html" по умолчанию
}

// Server представляет собой сервер для обслуживания статики.
type Server struct {
	cache        *Cache
	pb           *utils.PathBuilder
	fallbackPath string
}

// NewServer создает новый сервер для статики.
func NewServer(disableCache bool, rootPath string, fromLen, capHint int, fallbackName string) *Server {
	if fallbackName == "" {
		fallbackName = "index.html"
	}
	var cache *Cache
	if !disableCache {
		cache = NewCache()
	}
	return &Server{
		cache:        cache,
		pb:           utils.NewPathBuilder(rootPath, fromLen, capHint),
		fallbackPath: rootPath + fallbackName,
	}
}

// Handle запросы на файлы.
func (s *Server) Handle(ctx *fasthttp.RequestCtx) {
	fullPath := s.pb.Build(ctx.Path())
	entry := s.lookupFile(fullPath)
	if entry == nil {
		newEntry, code, err := LoadFile(fullPath)
		if err != nil {
			switch code {
			case fasthttp.StatusNotFound:
				newEntry, code, err = LoadFile(s.fallbackPath)
				if err != nil {
					ctx.Error(err.Error(), code)
					return
				}
			default:
				ctx.Error(err.Error(), code)
				return
			}
		}
		entry = newEntry
		if s.cache != nil {
			s.cache.Set(fullPath, entry)
		}
	}
	err := writeFileEntry(ctx, entry)
	if err != nil {
		log.Error().Err(err).Msg("ctx.Write")
		return
	}
}

func writeFileEntry(ctx *fasthttp.RequestCtx, fileEntry *File) error {
	ctx.Response.Header.Set("Content-Type", fileEntry.ContentType)
	ctx.Response.Header.Set("Content-Length", strconv.Itoa(len(fileEntry.Body)))
	ctx.Response.SetBodyRaw(fileEntry.Body)
	return nil
}

func (s *Server) lookupFile(filePath string) *File {
	if s.cache == nil {
		return nil
	}
	entry := s.cache.Get(filePath)
	switch {
	case entry == nil:
		return nil
	case entry.IsExpired():
		s.cache.Delete(filePath)
		return nil
	default:
		return entry
	}
}
