//nolint:wrapcheck
package run

import (
	"fmt"
	"time"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/internal/metrics"
	"go.xsfx.dev/schnutibox/pkg/rfid"
	"go.xsfx.dev/schnutibox/pkg/web"
)

const TickerTime = time.Second

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

func (m *mpc) play(logger zerolog.Logger, rfid string, name string, uris []string) error {
	logger.Info().Msg("trying to add tracks")

	// Stop playing track.
	if err := m.stop(logger); err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	// Clear playlist.
	if err := m.clear(logger); err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	// Adding every single uri to playlist
	for _, i := range uris {
		logger.Debug().Str("uri", i).Msg("add track")

		if err := m.conn.Add(i); err != nil {
			metrics.BoxErrors.Inc()

			return err
		}
	}

	// Getting playlist uris from MPD server.
	// This is needed to identify the right metric to use.
	mpdURIS, err := m.playlistURIS()
	if err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	metrics.NewPlay(rfid, name, mpdURIS)

	return m.conn.Play(-1)
}

// playlistURIS extracts uris from MPD playlist.
func (m *mpc) playlistURIS() ([]string, error) {
	// Check if we can connect to MPD server.
	if err := m.conn.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping MPD server: %w", err)
	}

	attrs, err := m.conn.PlaylistInfo(-1, -1)
	if err != nil {
		return nil, fmt.Errorf("could not get playlist: %w", err)
	}

	// Stores the tracklist it got from the MPD server.
	uris := []string{}

	// Builds uri list.
	for _, a := range attrs {
		uris = append(uris, a["file"])
	}

	return uris, nil
}

func (m *mpc) watch() {
	log.Debug().Msg("starting watch")

	ticker := time.NewTicker(TickerTime)

	go func() {
		for {
			<-ticker.C

			uris, err := m.playlistURIS()
			if err != nil {
				log.Error().Err(err).Msg("could not get playlist uris")
				metrics.BoxErrors.Inc()

				continue
			}

			// Gettings MPD state.
			s, err := m.conn.Status()
			if err != nil {
				log.Error().Err(err).Msg("could not get status")
				metrics.BoxErrors.Inc()

				continue
			}

			// Sets the metrics.
			metrics.Set(uris, s["state"])
		}
	}()
}

func Run(cmd *cobra.Command, args []string) {
	log.Info().Msg("starting the RFID reader")

	idChan := make(chan string)
	r := rfid.NewRFID(config.Cfg, idChan)

	if err := r.Run(); err != nil {
		log.Fatal().Err(err).Msg("could not start RFID reader")
	}

	go func() {
		var id string

		for {
			// Wating for a scanned tag.
			id = <-idChan
			logger := log.With().Str("id", id).Logger()
			logger.Info().Msg("received id")

			// Create MPD connection on every received event.
			c, err := mpd.Dial("tcp", fmt.Sprintf("%s:%d", config.Cfg.MPD.Hostname, config.Cfg.MPD.Port))
			if err != nil {
				log.Fatal().Err(err).Msg("could not connect to MPD server")
			}

			m := newMpc(c)

			// Check of stop tag was detected.
			if id == config.Cfg.Meta.Stop {
				logger.Info().Msg("stopping")

				if err := m.stop(logger); err != nil {
					logger.Error().Err(err).Msg("could not stop")
				}

				if err := m.clear(logger); err != nil {
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
			if err := m.play(logger, id, tracks.Name, tracks.Uris); err != nil {
				logger.Error().Err(err).Msg("could not play track")
			}
		}
	}()

	// Running web interface. Blocking.
	web.Run(cmd, args)
}
