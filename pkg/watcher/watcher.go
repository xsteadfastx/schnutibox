package watcher

import (
	"time"

	"github.com/rs/zerolog/log"
	"go.xsfx.dev/schnutibox/internal/metrics"
	"go.xsfx.dev/schnutibox/pkg/mpc"
	"go.xsfx.dev/schnutibox/pkg/timer"
)

const tickerTime = time.Second

// Run runs actions after tickerTime is over, over again and again.
// Right now its mostly used for setting metrics.
func Run() {
	log.Debug().Msg("starting watch")

	ticker := time.NewTicker(tickerTime)

	go func() {
		for {
			<-ticker.C

			// Timer.
			go timer.T.Handle()

			// Metrics.
			go func() {
				m, err := mpc.Conn()
				if err != nil {
					log.Error().Err(err).Msg("could not connect")

					return
				}

				uris, err := mpc.PlaylistURIS()
				if err != nil {
					log.Error().Err(err).Msg("could not get playlist uris")
					metrics.BoxErrors.Inc()

					return
				}

				// Gettings MPD state.
				s, err := m.Status()
				if err != nil {
					log.Error().Err(err).Msg("could not get status")
					metrics.BoxErrors.Inc()

					return
				}

				// Sets the metrics.
				metrics.Set(uris, s["state"])
			}()
		}
	}()
}
