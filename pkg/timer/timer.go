package timer

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
)

// nolint:gochecknoglobals
var T = &api.Timer{}

func Timer() {
	if T.Duration != nil {
		// Initialize the current object.
		if T.Current == nil {
			T.Current = &duration.Duration{}
			T.Current.Seconds = T.Duration.Seconds
		}

		switch {
		// There is some timing going on.
		case T.Duration.Seconds != 0 && T.Current.Seconds != 0:
			log.Debug().
				Int64("current", T.Current.Seconds).
				Int64("duration", T.Duration.Seconds).
				Msg("timer is running")

			if T.Current.Seconds > 0 {
				T.Current.Seconds -= 1

				return
			}

		// No timer is running... so setting the duration to 0.
		// TODO: Needs to do something actually!
		case T.Current.Seconds == 0 && T.Duration.Seconds != 0:
			log.Debug().Msg("stoping timer")

			T.Duration.Seconds = 0
		}
	}
}

func Run(cmd *cobra.Command, args []string) {}
