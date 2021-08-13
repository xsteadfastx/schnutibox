package main

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.xsfx.dev/schnutibox/cmd"
	"go.xsfx.dev/schnutibox/pkg/sselog"
)

func main() {
	sselog.Log = sselog.NewSSELog()
	log.Logger = zerolog.New(io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stderr}, sselog.Log)).With().Caller().Logger()

	cmd.Execute()
}
