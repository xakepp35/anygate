package handler

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/config"
	"github.com/xakepp35/anygate/plugin"
	"github.com/xakepp35/anygate/router"
)

// 🏗️ Register — чертёж памяти, где каждый путь знает свою судьбу.
func Register(r *router.Router, cfg config.Root, inheritedPlugins ...plugin.Spec) {
	// Склеиваем middleware цепочку текущий группы
	fullChain := make([]plugin.Spec, 0, len(inheritedPlugins)+len(cfg.Plugins))
	fullChain = append(fullChain, inheritedPlugins...)
	fullChain = append(fullChain, cfg.Plugins...)
	// Регистрируем маршруты текущей группы
	for fromSpec, to := range cfg.Routes {
		methods, from := parseFromSpec(fromSpec)
		base, mode := New(from, to, cfg)
		final := plugin.BuildChain(base, fullChain...)
		for _, method := range methods {
			registerRoute(r, method, from, final)
			log.Info().Str("from", from).Str("method", method).Str("to", to).Str("mode", mode).Msg("route")
		}
	}
	// Рекурсивно обрабатываем подгруппы
	for _, child := range cfg.Groups {
		Register(r, child, fullChain...)
	}
}

func New(from, to string, cfg config.Root) (fasthttp.RequestHandler, string) {
	switch {
	case to == "":
		return Ok, "ok"
	case to == "*":
		return Echo, "echo"
	case strings.HasPrefix(to, "http://") || strings.HasPrefix(to, "https://"):
		return NewProxy(from, to, cfg.Proxy), "proxy"
	case isFixedResponse(to):
		code, body := parseFixedResponse(to)
		return NewFixed(code, body), "fixed"
	default:
		return NewStatic(from, to, cfg.Static), "static"
	}
}

func parseFromSpec(spec string) ([]string, string) {
	parts := strings.Fields(spec)
	if len(parts) < 2 {
		return []string{"ANY"}, spec // если нет пробелов — просто путь
	}
	methods := make([]string, 0, len(parts)-1)
	for _, mstr := range parts[:len(parts)-1] {
		m := strings.ToUpper(mstr)
		if m == "*" || m == "ANY" {
			return []string{"ANY"}, parts[len(parts)-1]
		}
		methods = append(methods, strings.ToUpper(m))
	}
	return methods, parts[len(parts)-1]
}

// func parseFromSpec(to string) (method, path string) {
// 	for i := 0; i < len(to); i++ {
// 		if to[i] == ' ' {
// 			return strings.ToUpper(to[:i]), to[i+1:]
// 		}
// 	}
// 	return "ANY", to
// }

func registerRoute(r *router.Router, method, path string, h fasthttp.RequestHandler) {
	r.Register(method, path, h)
}

// func registerRoute(r *router.Router, method, path string, h fasthttp.RequestHandler) {
// 	switch method {
// 	case "ANY":
// 		r.ANY(path, h)
// 	case "GET":
// 		r.GET(path, h)
// 	case "HEAD":
// 		r.HEAD(path, h)
// 	case "POST":
// 		r.POST(path, h)
// 	case "PUT":
// 		r.PUT(path, h)
// 	case "PATCH":
// 		r.PATCH(path, h)
// 	case "DELETE":
// 		r.DELETE(path, h)
// 	case "CONNECT":
// 		r.CONNECT(path, h)
// 	case "OPTIONS":
// 		r.OPTIONS(path, h)
// 	case "TRACE":
// 		r.TRACE(path, h)
// 	default:
// 		log.Fatal().Str("method", method).Str("path", path).Msg("unsupported method")
// 	}
// }
