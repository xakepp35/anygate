package proxy

import "github.com/valyala/fasthttp"

var DefaultProxyConfig = ProxyConfig{
	Timeout:              0,     // infinite
	InsecureSkipVerify:   false, // secure
	RouteLenHint:         64,
	StatusBadGateway:     fasthttp.StatusBadGateway,
	StatusGatewayTimeout: fasthttp.StatusGatewayTimeout,
}

func (inherited ProxyConfigOpt) render() ProxyConfig {
	rendered := DefaultProxyConfig
	switch {
	case inherited.Timeout != nil:
		rendered.Timeout = *inherited.Timeout
	case inherited.InsecureSkipVerify != nil:
		rendered.InsecureSkipVerify = *inherited.InsecureSkipVerify
	case inherited.RouteLenHint != nil:
		rendered.RouteLenHint = *inherited.RouteLenHint
	case inherited.StatusBadGateway != nil:
		rendered.StatusBadGateway = *inherited.StatusBadGateway
	case inherited.StatusGatewayTimeout != nil:
		rendered.StatusGatewayTimeout = *inherited.StatusGatewayTimeout
	}
	return rendered
}

// Наследуем конфиг с оверрайдами
func (inherited ProxyConfigOpt) override(group ProxyConfigOpt) ProxyConfigOpt {
	switch {
	case group.Timeout != nil:
		inherited.Timeout = group.Timeout
	case group.InsecureSkipVerify != nil:
		inherited.InsecureSkipVerify = group.InsecureSkipVerify
	case group.RouteLenHint != nil:
		inherited.RouteLenHint = group.RouteLenHint
	case group.StatusBadGateway != nil:
		inherited.StatusBadGateway = group.StatusBadGateway
	case group.StatusGatewayTimeout != nil:
		inherited.StatusGatewayTimeout = group.StatusGatewayTimeout
	}
	return inherited
}
