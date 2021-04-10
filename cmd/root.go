//nolint:exhaustivestruct,gochecknoglobals,gochecknoinits
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/pkg/prepare"
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
	// Run.
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file")

	if err := runCmd.MarkFlagRequired("config"); err != nil {
		log.Fatal().Err(err).Msg("missing flag")
	}

	// Prepare.
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.Flags().BoolVar(&prepare.Cfg.ReadOnly, "read-only", false, "Setup read-only system")
	prepareCmd.Flags().StringVarP(&prepare.Cfg.System, "system", "s", "raspbian", "Which kind of system to prepare")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyUsername, "spotify-username", "", "Spotify username")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyPassword, "spotify-password", "", "Spotify password")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyClientID, "spotify-client-id", "", "Spotify client ID")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyClientSecret, "spotify-client-secret", "", "Spotify client secret")

	// Version.
	rootCmd.AddCommand(versionCmd)
}

// initConfig loads the config file.
func initConfig() {
	logger := log.With().Str("config", cfgFile).Logger()

	// Defaults.
	viper.SetDefault("box.hostname", "localhost")
	viper.SetDefault("box.port", 9999)
	viper.SetDefault("mpd.hostname", "localhost")
	viper.SetDefault("mpd.port", 6600)
	viper.SetDefault("reader.dev", "/dev/hidraw0")

	// Environment handling.
	viper.SetEnvPrefix("SCHNUTIBOX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Parse config file.
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		parseConfig(logger)
	} else {
		logger.Fatal().Msg("missing config file")
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Info().Msg("config file changed")
		parseConfig(logger)
	})
}

func parseConfig(logger zerolog.Logger) {
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal().Err(err).Msg("error loading config file")
	}

	if err := viper.Unmarshal(&config.Cfg); err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal config")
	}

	if err := config.Cfg.Require(); err != nil {
		logger.Fatal().Err(err).Msg("missing config parts")
	}
}

// Execute executes the commandline interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
