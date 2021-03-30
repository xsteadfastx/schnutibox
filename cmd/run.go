package cmd

import (
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/run"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Running this thing",
	Run:   run.Run,
}
