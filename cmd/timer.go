package cmd

import (
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/timer"
)

// nolint:gochecknoglobals
var timerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Handling timer",
	Run:   timer.Run,
}
