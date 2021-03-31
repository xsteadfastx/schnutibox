//nolint:exhaustivestruct,gochecknoglobals
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TracksPlayed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "schnutibox_played_tracks_total",
		},
		[]string{"rfid", "name"})

	BoxErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "schnutbox_errors_total",
		},
	)
)
