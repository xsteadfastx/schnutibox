package run

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/pkg/mpc"
	"go.xsfx.dev/schnutibox/pkg/rfid"
	"go.xsfx.dev/schnutibox/pkg/watcher"
	"go.xsfx.dev/schnutibox/pkg/web"
)

func Run(cmd *cobra.Command, args []string) {
	log.Info().Msg("starting the RFID reader")

	idChan := make(chan string)
	r := rfid.NewRFID(config.Cfg, idChan)

	if err := r.Run(); err != nil {
		if !viper.GetBool("reader.ignore") {
			log.Fatal().Err(err).Msg("could not start RFID reader")
		}

		log.Warn().Err(err).Msg("could not start RFID reader. ignoring...")
	}

	// Stating watcher.
	watcher.Run()

	// nolint:nestif
	if !viper.GetBool("reader.ignore") {
		go func() {
			var id string

			for {
				// Wating for a scanned tag.
				id = <-idChan
				logger := log.With().Str("id", id).Logger()
				logger.Info().Msg("received id")

				// Check of stop tag was detected.
				if id == config.Cfg.Meta.Stop {
					logger.Info().Msg("stopping")

					if err := mpc.Stop(logger); err != nil {
						logger.Error().Err(err).Msg("could not stop")
					}

					if err := mpc.Clear(logger); err != nil {
						logger.Error().Err(err).Msg("could not clear")
					}

					continue
				}

				// Check if there is a track for the ID.
				tracks, ok := config.Cfg.Tracks[id]
				if !ok {
					logger.Error().Msg("could not find track for ID")

					continue
				}

				// Try to play track.
				if err := mpc.Play(logger, id, tracks.Name, tracks.Uris); err != nil {
					logger.Error().Err(err).Msg("could not play track")
				}
			}
		}()
	}

	// Running web interface. Blocking.
	web.Run(cmd, args)
}
