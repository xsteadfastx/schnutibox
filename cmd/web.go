// nolint:gochecknoglobals
package cmd

import (
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/web"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starting webservice",
	Run:   web.Run,
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig(false)
	},
}
