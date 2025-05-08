package main

import (
	"github.com/rs/zerolog/log"
	"github.com/xakepp35/pkg/xlog"
)

func init() {
	xlog.Init()
}

func main() {
	log.Fatal().Msg("not implemented")
}
