package main

import (
	"os"

	"github.com/fasthttp/router"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/plugin"
	"github.com/xakepp35/anygate/proxy"
	"github.com/xakepp35/pkg/xlog"
	"gopkg.in/yaml.v3"
)

func init() {
	xlog.Init()
}

func main() {
	// 🚀 Загружаем конфиг
	configPath := "/config.yml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", configPath).Msg("os.ReadFile")
	}

	var group proxy.RoutesGroup
	if err := yaml.Unmarshal(data, &group); err != nil {
		log.Fatal().Err(err).Msg("yaml.Unmarshal")
	}

	// 🧠 Инициализируем роутер и проксируем
	r := router.New()
	proxy.BuildGroup(r, group, proxy.ProxyConfigOpt{}, []plugin.Spec{}...)

	log.Info().Msg("AnyGate listening at :8000")
	if err := fasthttp.ListenAndServe(":8000", r.Handler); err != nil {
		log.Fatal().Err(err).Msg("fasthttp.ListenAndServe")
	}
}
