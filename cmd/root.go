//nolint:exhaustivestruct,gochecknoglobals
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.xsfx.dev/schnutibox/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "schnutibox",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Usage(); err != nil {
			log.Error().Msg(err.Error())
		}
	},
}

// init initializes the command line interface.
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(runCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	// Defaults
	viper.SetDefault("box.hostname", "localhost")
	viper.SetDefault("box.port", 9999)
	viper.SetDefault("mpd.hostname", "localhost")
	viper.SetDefault("mpd.port", 6600)
	viper.SetDefault("reader.dev", "/dev/hidraw0")
}

// initConfig loads the config file.
// TODO: needs some environment variable love!
func initConfig() {
	logger := log.With().Str("config", cfgFile).Logger()
	viper.SetEnvPrefix("SCHNUTIBOX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			logger.Fatal().Err(err).Msg("error loading config file")
		}
		if err := viper.Unmarshal(&config.Cfg); err != nil {
			logger.Fatal().Err(err).Msg("could not unmarshal config")
		}
		if err := config.Cfg.Require(); err != nil {
			logger.Fatal().Err(err).Msg("missing config parts")
		}
	} else {
		logger.Fatal().Msg("missing config file")
	}
}

// Execute executes the commandline interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
