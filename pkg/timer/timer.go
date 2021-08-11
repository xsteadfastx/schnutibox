package timer

import (
	"context"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/internal/grpcclient"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
	"go.xsfx.dev/schnutibox/pkg/mpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

// nolint:gochecknoglobals
var T = &Timer{}

type Timer struct {
	Req *api.Timer
}

func (t *Timer) Handle() {
	if t.Req != nil {
		// Initialize the current object.
		if t.Req.Current == nil {
			t.Req.Current = &duration.Duration{}
			t.Req.Current.Seconds = t.Req.Duration.Seconds
		}

		switch {
		// There is some timing going on.
		case t.Req.Duration.Seconds != 0 && t.Req.Current.Seconds != 0:
			log.Debug().
				Int64("current", t.Req.Current.Seconds).
				Int64("duration", t.Req.Duration.Seconds).
				Msg("timer is running")

			if t.Req.Current.Seconds > 0 {
				t.Req.Current.Seconds -= 1

				return
			}

		// No timer is running... so setting the duration to 0.
		case t.Req.Current.Seconds == 0 && t.Req.Duration.Seconds != 0:
			log.Debug().Msg("stoping timer")

			if err := mpc.Stop(log.Logger); err != nil {
				log.Error().Err(err).Msg("could not stop")
			}

			t.Req.Duration.Seconds = 0
		}
	}
}

// Run is the command line interface for triggering the timer.
func Run(cmd *cobra.Command, args []string) {
	conn, err := grpcclient.Conn(config.Cfg.Web.Hostname, config.Cfg.Web.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect")
	}

	c := api.NewTimerServiceClient(conn)

	d := durationpb.New(viper.GetDuration("timer.duration"))

	_, err = c.Create(context.Background(), &api.Timer{Duration: d})
	if err != nil {
		conn.Close()
		log.Fatal().Err(err).Msg("could not create timer")
	}

	conn.Close()
	log.Info().Msg("added timer")
}
