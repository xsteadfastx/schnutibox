//nolint:gochecknoglobals
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		// nolint: forbidigo
		fmt.Printf("schnutibox %s, commit %s, %s", version, commit, date)
	},
}
