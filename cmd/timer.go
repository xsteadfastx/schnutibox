package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/timer"
)

// nolint:gochecknoglobals
var timerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Handling timer",
	Run:   timer.Run,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		initConfig(false)
	},
}
