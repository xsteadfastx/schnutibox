//nolint:exhaustivestruct,gochecknoglobals,gochecknoinits,gomnd
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
//
// nolint:funlen
func init() {
	// Root.
	rootCmd.PersistentFlags().Bool("pprof", false, "Enables pprof for debugging")

	// Run.
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file")

	if err := runCmd.MarkFlagRequired("config"); err != nil {
		log.Fatal().Err(err).Msg("missing flag")
	}

	runCmd.Flags().Bool("ignore-reader", false, "Ignoring that the reader is missing")

	// Prepare.
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.Flags().BoolVar(&prepare.Cfg.ReadOnly, "read-only", false, "Setup read-only system")
	prepareCmd.Flags().StringVarP(&prepare.Cfg.System, "system", "s", "raspbian", "Which kind of system to prepare")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyUsername, "spotify-username", "", "Spotify username")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyPassword, "spotify-password", "", "Spotify password")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyClientID, "spotify-client-id", "", "Spotify client ID")
	prepareCmd.Flags().StringVar(&prepare.Cfg.SpotifyClientSecret, "spotify-client-secret", "", "Spotify client secret")
	prepareCmd.Flags().StringVar(&prepare.Cfg.RFIDReader, "rfid-reader", "/dev/hidraw0", "dev path of rfid reader")
	prepareCmd.Flags().StringVar(&prepare.Cfg.StopID, "stop-id", "", "ID of stop tag")

	// Version.
	rootCmd.AddCommand(versionCmd)

	// Web.
	rootCmd.AddCommand(webCmd)
	webCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file")

	if err := webCmd.MarkFlagRequired("config"); err != nil {
		log.Fatal().Err(err).Msg("missing flag")
	}

	// Timer.
	rootCmd.AddCommand(timerCmd)
	timerCmd.Flags().String("hostname", "localhost", "Hostname of schnutibox")
	timerCmd.Flags().Int("port", 6600, "Port of schnutibox")
	timerCmd.Flags().DurationP("duration", "d", time.Minute, "Duration until the timer stops the playback")

	if err := timerCmd.MarkFlagRequired("duration"); err != nil {
		log.Fatal().Err(err).Msg("missing flag")
	}

	// Defaults.
	viper.SetDefault("web.hostname", "localhost")
	viper.SetDefault("web.port", 9999)
	viper.SetDefault("mpd.hostname", "localhost")
	viper.SetDefault("mpd.port", 6600)
	viper.SetDefault("reader.dev", "/dev/hidraw0")
	viper.SetDefault("reader.ignore", false)
	viper.SetDefault("debug.pprof", false)
	viper.SetDefault("timer.duration", time.Minute)

	// Environment handling.
	viper.SetEnvPrefix("SCHNUTIBOX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Flags.
	for k, v := range map[string]*pflag.Flag{
		"debug.pprof":    rootCmd.PersistentFlags().Lookup("pprof"),
		"reader.dev":     prepareCmd.Flags().Lookup("rfid-reader"),
		"reader.ignore":  runCmd.Flags().Lookup("ignore-reader"),
		"web.hostname":   timerCmd.Flags().Lookup("hostname"),
		"web.port":       timerCmd.Flags().Lookup("port"),
		"timer.duration": timerCmd.Flags().Lookup("duration"),
	} {
		if err := viper.BindPFlag(k, v); err != nil {
			log.Fatal().Err(err).Msg("could not bind flag")
		}
	}
}

// initConfig loads the config file.
// fatal defines if config parsing should end in a fatal error or not.
func initConfig(fatal bool) {
	logger := log.With().Str("config", cfgFile).Logger()

	// Parse config file.
	if cfgFile == "" && fatal {
		logger.Fatal().Msg("missing config file")
	} else if cfgFile == "" {
		logger.Warn().Msg("missing config file")
	}

	// Dont mind if there is no config file... viper also should populate
	// flags and environment variables.
	viper.SetConfigFile(cfgFile)
	parseConfig(logger, fatal)

	// Configfile changes watch only enabled if there is a config file.
	if cfgFile != "" {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			logger.Info().Msg("config file changed")
			parseConfig(logger, false)
		})
	}
}

// parseConfig parses the config and does some tests if required fields are there.
// Its also possible to decide if parsing should end up in a fatal or just an error.
func parseConfig(logger zerolog.Logger, fatal bool) {
	if err := viper.ReadInConfig(); err != nil {
		if fatal {
			logger.Fatal().Err(err).Msg("error loading config file")
		}

		logger.Error().Err(err).Msg("error loading config file")
	}

	if err := viper.Unmarshal(&config.Cfg); err != nil {
		if fatal {
			logger.Fatal().Err(err).Msg("could not unmarshal config")
		}

		logger.Error().Err(err).Msg("could not unmarshal config")
	}

	// Disabling require check if no config is set.
	// Not sure about this!
	if cfgFile != "" {
		if err := config.Cfg.Require(); err != nil {
			if fatal {
				logger.Fatal().Err(err).Msg("missing config parts")
			}

			logger.Error().Err(err).Msg("missing config parts")

			return
		}
	} else {
		logger.Warn().Msg("doesnt do a config requirement check")

		return
	}
}

// Execute executes the commandline interface.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
