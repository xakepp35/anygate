package proxy

import (
	"strings"

	"github.com/fasthttp/router"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/plugin"
)

// 🏗️ BuildGroup — чертёж памяти, где каждый путь знает свою судьбу.
func BuildGroup(r *router.Router, group RoutesGroup, inheritedConfig ProxyConfigOpt, inheritedPlugins ...plugin.Spec) {
	overriddenCfg := inheritedConfig.override(group.Config)
	renderedCfg := overriddenCfg.render()
	// Склеиваем middleware цепочку текущий группы
	fullChain := append([]plugin.Spec{}, inheritedPlugins...)
	fullChain = append(fullChain, group.Plugins...)
	// Регистрируем маршруты текущей группы
	for from, to := range group.Routes {
		method, target := parseTargetSpec(to)
		base := Handler(from, target, renderedCfg)
		final := plugin.BuildChain(base, fullChain...)
		registerRoute(r, method, from, final)
		log.Info().Str("from", from).Str("method", method).Str("target", target).Msg("proxy")
	}
	// Рекурсивно обрабатываем подгруппы
	for _, child := range group.Groups {
		BuildGroup(r, child, overriddenCfg, fullChain...)
	}
}

func parseTargetSpec(to string) (method, target string) {
	parts := strings.SplitN(to, " ", 2)
	if len(parts) == 2 {
		return strings.ToUpper(parts[0]), parts[1]
	}
	return "ANY", to
}

func registerRoute(r *router.Router, method, path string, h fasthttp.RequestHandler) {
	switch method {
	case "ANY":
		r.ANY(path, h)
	case "GET":
		r.GET(path, h)
	case "HEAD":
		r.HEAD(path, h)
	case "POST":
		r.POST(path, h)
	case "PUT":
		r.PUT(path, h)
	case "PATCH":
		r.PATCH(path, h)
	case "DELETE":
		r.DELETE(path, h)
	case "CONNECT":
		r.CONNECT(path, h)
	case "OPTIONS":
		r.OPTIONS(path, h)
	case "TRACE":
		r.TRACE(path, h)
	default:
		log.Fatal().Str("method", method).Str("path", path).Msg("unsupported method")
	}
}
