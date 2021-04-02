//nolint:exhaustivestruct
package prepare

import (
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	systemPrompt := promptui.Select{
		Label: "What kind of system to prepare",
		Items: []string{"Raspbian Lite"},
	}

	_, system, err := systemPrompt.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("could not get system")
	}

	log.Print(system)
}
