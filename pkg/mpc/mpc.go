package mpc

import (
	"errors"
	"fmt"
	"time"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/internal/metrics"
)

const (
	TickerTime  = time.Second
	Timeout     = 5 * time.Second
	TimeoutWait = time.Second / 2
)

var (
	ErrCouldNotConnect = errors.New("could not connect")
	ErrTimeout         = errors.New("timeout")
)

func Conn() (*mpd.Client, error) {
	t := time.NewTimer(Timeout)

	for {
		select {
		case <-t.C:
			return nil, ErrTimeout
		default:
			c, err := mpd.Dial("tcp", fmt.Sprintf("%s:%d", config.Cfg.MPD.Hostname, config.Cfg.MPD.Port))
			if err != nil {
				log.Error().Err(err).Msg("could not connect")

				time.Sleep(TimeoutWait)

				continue
			}

			if !t.Stop() {
				go func() {
					<-t.C
				}()
			}

			return c, nil
		}
	}
}

// PlaylistURIS extracts uris from MPD playlist.
func PlaylistURIS() ([]string, error) {
	m, err := Conn()
	if err != nil {
		return nil, fmt.Errorf("could not connect to MPD server :%w", err)
	}

	attrs, err := m.PlaylistInfo(-1, -1)
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

func Watcher() {
	log.Debug().Msg("starting watch")

	ticker := time.NewTicker(TickerTime)

	go func() {
		for {
			<-ticker.C

			m, err := Conn()
			if err != nil {
				log.Error().Err(err).Msg("could not connect")

				continue
			}

			uris, err := PlaylistURIS()
			if err != nil {
				log.Error().Err(err).Msg("could not get playlist uris")
				metrics.BoxErrors.Inc()

				continue
			}

			// Gettings MPD state.
			s, err := m.Status()
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

func Stop(logger zerolog.Logger) error {
	logger.Info().Msg("trying to stop playback")

	m, err := Conn()
	if err != nil {
		return fmt.Errorf("could not connect: %w", err)
	}

	return m.Stop()
}

func Clear(logger zerolog.Logger) error {
	logger.Info().Msg("trying to clear playlist")

	m, err := Conn()
	if err != nil {
		return fmt.Errorf("could not connect: %w", err)
	}

	return m.Clear()
}

func Play(logger zerolog.Logger, rfid string, name string, uris []string) error {
	logger.Info().Msg("trying to add tracks")

	m, err := Conn()
	if err != nil {
		return fmt.Errorf("could not connect: %w", err)
	}

	// Stop playing track.
	if err := Stop(logger); err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	// Clear playlist.
	if err := Clear(logger); err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	// Adding every single uri to playlist
	for _, i := range uris {
		logger.Debug().Str("uri", i).Msg("add track")

		if err := m.Add(i); err != nil {
			metrics.BoxErrors.Inc()

			return err
		}
	}

	// Getting playlist uris from MPD server.
	// This is needed to identify the right metric to use.
	mpdURIS, err := PlaylistURIS()
	if err != nil {
		metrics.BoxErrors.Inc()

		return err
	}

	metrics.NewPlay(rfid, name, mpdURIS)

	return m.Play(-1)
}
