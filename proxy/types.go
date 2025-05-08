package proxy

import (
	"time"

	"github.com/xakepp35/anygate/plugin"
)

type ProxyConfigOpt struct {
	Timeout              *time.Duration `yaml:"timeout"`              // infinite by default
	InsecureSkipVerify   *bool          `yaml:"insecureSkipVerify"`   // set true to ignore invalid https certs
	RouteLenHint         *int           `yaml:"routeLenHint"`         // 64 by default, set more if you use wider get queries
	StatusBadGateway     *int           `yaml:"statusBadGateway"`     // 502 by default
	StatusGatewayTimeout *int           `yaml:"statusGatewayTimeout"` // 504 by default
}

type ProxyConfig struct {
	Timeout              time.Duration `yaml:"timeout"`              // infinite by default
	InsecureSkipVerify   bool          `yaml:"insecureSkipVerify"`   // set true to ignore invalid https certs
	RouteLenHint         int           `yaml:"routeLenHint"`         // 64 by default, set more if you use wider get queries
	StatusBadGateway     int           `yaml:"statusBadGateway"`     // 502 by default
	StatusGatewayTimeout int           `yaml:"statusGatewayTimeout"` // 504 by default
}

type RoutesGroup struct {
	Config  ProxyConfigOpt    `yaml:"config"`
	Routes  map[string]string `yaml:"routes"`
	Plugins []plugin.Spec     `yaml:"plugins"`
	Groups  []RoutesGroup     `yaml:"groups"`
}
