package handler

import (
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/config"
)

// 🚀 NewProxy — минималистичный шприц, что вшивает внешний мир в твой процесс, избегая копий, но сохраняя смысл.
func NewProxy(from, to string, cfg config.Proxy) fasthttp.RequestHandler {
	if cfg.StatusBadGateway == 0 {
		cfg.StatusBadGateway = fasthttp.StatusBadGateway
	}
	if cfg.StatusGatewayTimeout == 0 {
		cfg.StatusGatewayTimeout = fasthttp.StatusGatewayTimeout
	}
	if cfg.RouteLenHint == 0 {
		cfg.RouteLenHint = 64
	}
	builderPool := &sync.Pool{
		New: func() any {
			b := strings.Builder{}
			b.Grow(cfg.RouteLenHint)
			return &b
		},
	}
	client := NewClient(cfg)
	// closure workload - caution, hot path!
	return func(ctx *fasthttp.RequestCtx) {
		upstreamBuilder := builderPool.Get().(*strings.Builder)
		upstreamBuilder.Reset()
		upstreamBuilder.WriteString(to)
		upstreamBuilder.Write(ctx.Path()[len(from):])
		ctx.Request.SetRequestURI(upstreamBuilder.String())
		builderPool.Put(upstreamBuilder)
		ctx.Request.SetTimeout(cfg.Timeout)
		err := client.Do(&ctx.Request, &ctx.Response)
		switch err {
		case nil:
			// NO ERROR -> DO NOTHING
		case fasthttp.ErrTimeout:
			ctx.SetStatusCode(cfg.StatusGatewayTimeout)
			ctx.SetBodyString(`{"error":"timeout"}`)
			log.Error().Err(err).Str("to", to).Str("from", from).Msg("timeout")
		default:
			ctx.SetStatusCode(cfg.StatusBadGateway)
			ctx.SetBodyString(`{"error":"` + err.Error() + `"}`)
			log.Error().Err(err).Str("to", to).Str("from", from).Msg("gateway")
		}
	}
}
