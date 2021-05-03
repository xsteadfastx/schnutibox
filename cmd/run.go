//nolint:gochecknoglobals
package cmd

import (
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/run"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Running this thing",
	Run:   run.Run,
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig(true)
	},
}
