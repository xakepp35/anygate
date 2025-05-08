package plugin

import (
	"errors"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

func PluginRateLimiter(args SpecArgs) (Func, error) {
	rawRPS, ok := args["rps"]
	if !ok {
		return nil, errors.New("missing 'rps'")
	}
	rps, ok := rawRPS.(float64)
	if !ok {
		return nil, errors.New("'rps' must be number")
	}

	burst := int(rps)
	if b, ok := args["burst"].(float64); ok {
		burst = int(b)
	}
	// конструируем limiter один раз
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if !limiter.Allow() {
				ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
				ctx.SetBodyString(`{"error":"rate limit exceeded"}`)
				return
			}
			next(ctx)
		}
	}, nil
}
