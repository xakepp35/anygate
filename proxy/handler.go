package proxy

import (
	"crypto/tls"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
)

// 🚀 Handler — минималистичный шприц, что вшивает внешний мир в твой процесс, избегая копий, но сохраняя смысл.
func Handler(prefix, target string, cfg ProxyConfig) fasthttp.RequestHandler {
	builderPool := &sync.Pool{
		New: func() any {
			b := strings.Builder{}
			b.Grow(cfg.RouteLenHint)
			return &b
		},
	}
	client := fasthttp.Client{}
	if cfg.InsecureSkipVerify {
		client.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	if cfg.StatusBadGateway == 0 {
		cfg.StatusBadGateway = fasthttp.StatusBadGateway
	}
	if cfg.StatusGatewayTimeout == 0 {
		cfg.StatusGatewayTimeout = fasthttp.StatusGatewayTimeout
	}
	// closure workload - caution, hot path!
	return func(ctx *fasthttp.RequestCtx) {
		upstreamBuilder := builderPool.Get().(*strings.Builder)
		upstreamBuilder.Reset()
		upstreamBuilder.WriteString(target)
		upstreamBuilder.Write(ctx.Path()[len(prefix):])
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
		default:
			ctx.SetStatusCode(cfg.StatusBadGateway)
			ctx.SetBodyString(`{"error":"` + err.Error() + `"}`)
		}
	}
}
