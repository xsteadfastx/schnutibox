//nolint:wrapcheck
package run

import (
	"fmt"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/pkg/rfid"
)

type mpc struct {
	conn *mpd.Client
}

func newMpc(conn *mpd.Client) *mpc {
	return &mpc{conn}
}

func (m *mpc) stop(logger zerolog.Logger) error {
	logger.Info().Msg("trying to stop playback")

	return m.conn.Stop()
}

func (m *mpc) clear(logger zerolog.Logger) error {
	logger.Info().Msg("trying to clear playlist")

	return m.conn.Clear()
}

func (m *mpc) play(logger zerolog.Logger, uris []string) error {
	logger.Info().Msg("trying to add tracks")

	// Stop playing track.
	if err := m.stop(logger); err != nil {
		return err
	}

	// Clear playlist.
	if err := m.clear(logger); err != nil {
		return err
	}

	// Adding every single uri to playlist
	for _, i := range uris {
		logger.Debug().Str("uri", i).Msg("add track")

		if err := m.conn.Add(i); err != nil {
			return err
		}
	}

	return m.conn.Play(-1)
}

func Run(cmd *cobra.Command, args []string) {
	log.Info().Msg("starting the RFID reader")

	idChan := make(chan string)
	r := rfid.NewRFID(config.Cfg, idChan)

	if err := r.Run(); err != nil {
		log.Fatal().Err(err).Msg("could not start RFID reader")
	}

	var id string

	for {
		// Wating for a scanned tag.
		id = <-idChan
		logger := log.With().Str("id", id).Logger()
		logger.Info().Msg("received id")

		// Create MPD connection on every received event.
		c, err := mpd.Dial("tcp", fmt.Sprintf("%s:%d", config.Cfg.MPD.Hostname, config.Cfg.MPD.Port))
		if err != nil {
			logger.Error().Err(err).Msg("could not connect to MPD server")
		}

		m := newMpc(c)

		// Check of stop tag was detected.
		if id == config.Cfg.Meta.Stop {
			logger.Info().Msg("stopping")

			if err := m.stop(logger); err != nil {
				logger.Error().Err(err).Msg("could not stop")
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
		if err := m.play(logger, tracks); err != nil {
			logger.Error().Err(err).Msg("could not play track")
		}
	}
}
