package handler

import (
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/config"
)

// 🧠 NewServer apply config
func NewServer(cfg config.Server, handler fasthttp.RequestHandler) *fasthttp.Server {
	return &fasthttp.Server{
		Handler:                            handler,
		Name:                               cfg.Name,
		Concurrency:                        cfg.Concurrency,
		ReadBufferSize:                     cfg.ReadBufferSize,
		WriteBufferSize:                    cfg.WriteBufferSize,
		ReadTimeout:                        cfg.ReadTimeout,
		WriteTimeout:                       cfg.WriteTimeout,
		IdleTimeout:                        cfg.IdleTimeout,
		MaxConnsPerIP:                      cfg.MaxConnsPerIP,
		MaxRequestsPerConn:                 cfg.MaxRequestsPerConn,
		MaxIdleWorkerDuration:              cfg.MaxIdleWorkerDuration,
		TCPKeepalivePeriod:                 cfg.TCPKeepalivePeriod,
		MaxRequestBodySize:                 cfg.MaxRequestBodySize,
		SleepWhenConcurrencyLimitsExceeded: cfg.SleepOnConnLimitExceeded,
		DisableKeepalive:                   cfg.DisableKeepalive,
		TCPKeepalive:                       cfg.TCPKeepalive,
		ReduceMemoryUsage:                  cfg.ReduceMemoryUsage,
		GetOnly:                            cfg.GetOnly,
		DisablePreParseMultipartForm:       cfg.DisableMultipartForm,
		LogAllErrors:                       cfg.LogAllErrors,
		SecureErrorLogMessage:              cfg.SecureErrorLogMessage,
		DisableHeaderNamesNormalizing:      cfg.DisableHeaderNormalization,
		NoDefaultServerHeader:              cfg.NoDefaultServerHeader,
		NoDefaultDate:                      cfg.NoDefaultDate,
		NoDefaultContentType:               cfg.NoDefaultContentType,
		KeepHijackedConns:                  cfg.KeepHijackedConns,
		CloseOnShutdown:                    cfg.CloseOnShutdown,
		StreamRequestBody:                  cfg.StreamRequestBody,
	}
}
