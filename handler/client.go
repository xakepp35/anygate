package handler

import (
	"crypto/tls"

	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/config"
)

func NewClient(cfg config.Proxy) *fasthttp.Client {
	if cfg.Name == "" {
		cfg.Name = "anygate"
	}
	strategy := fasthttp.FIFO
	if cfg.ConnPoolStrategyLifo {
		strategy = fasthttp.LIFO
	}
	var tlsConfig *tls.Config
	if cfg.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return &fasthttp.Client{
		Name:                          cfg.Name,
		MaxConnsPerHost:               cfg.MaxConnsPerHost,
		MaxIdleConnDuration:           cfg.MaxIdleConnDuration,
		MaxConnDuration:               cfg.MaxConnDuration,
		MaxIdemponentCallAttempts:     cfg.MaxIdemponentCallAttempts,
		ReadBufferSize:                cfg.ReadBufferSize,
		WriteBufferSize:               cfg.WriteBufferSize,
		ReadTimeout:                   cfg.ReadTimeout,
		WriteTimeout:                  cfg.WriteTimeout,
		MaxResponseBodySize:           cfg.MaxResponseBodySize,
		MaxConnWaitTimeout:            cfg.MaxConnWaitTimeout,
		ConnPoolStrategy:              strategy,
		NoDefaultUserAgentHeader:      cfg.NoDefaultUserAgentHeader,
		DialDualStack:                 cfg.DialDualStack,
		DisableHeaderNamesNormalizing: cfg.DisableHeaderNamesNormalizing,
		DisablePathNormalizing:        cfg.DisablePathNormalizing,
		StreamResponseBody:            cfg.StreamResponseBody,
		TLSConfig:                     tlsConfig,
	}
}
