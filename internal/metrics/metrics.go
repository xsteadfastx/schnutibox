//nolint:exhaustivestruct,gochecknoglobals
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
)

// Plays is a map of tracked plays.
// Its a map, so its easier to check if the metric is already initialized
// and usable. The Key string is the RFID identification.
var Plays = make(map[string]*api.IdentifyResponse)

// NewPlay initialize a new play metric.
func NewPlay(rfid, name string, uris []string) {
	if _, ok := Plays[rfid]; !ok {
		Plays[rfid] = &api.IdentifyResponse{
			Name: name,
			Uris: uris,
		}
	}
}

// Play is the play metric.
var Play = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "schnutibox_plays",
		Help: "play metrics",
	},
	[]string{"rfid", "name"},
)

// BoxErrors counts schnutibox errors.
var BoxErrors = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "schnutibox_errors_total",
		Help: "counter of errors",
	},
)

// tracksEqual checks if uris slices are equal.
// This is needed to search for the right play item.
func tracksEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// Set sets `1` on play gauge if item is playing, a `0` on every other play.
func Set(uris []string, state string) {
	for r, p := range Plays {
		if tracksEqual(uris, p.Uris) && state == "play" {
			Play.WithLabelValues(r, p.Name).Set(1)
		} else {
			Play.WithLabelValues(r, p.Name).Set(0)
		}
	}
}
