package cmd

import (
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/pkg/prepare"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepares a device",
	Run:   prepare.Run,
}
