package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const defaultName = "anygate.yml"

// 🚀 NewRoot Загружает конфиг
func NewRoot() Root {
	// Можно оверрайднуть путь из CLI
	configPath := defaultName
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	// Читаем файл
	configPayload, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", configPath).Msg("os.ReadFile")
	}
	// Парсим yaml
	var cfg Root
	err = yaml.Unmarshal(configPayload, &cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("yaml.Unmarshal")
	}
	// Дефолты
	if cfg.Server.Name == "" {
		cfg.Server.Name = "anygate"
	}
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":80"
	}
	if cfg.Server.ListenNetwork == "" {
		cfg.Server.ListenNetwork = "tcp4"
	}
	return cfg
}
