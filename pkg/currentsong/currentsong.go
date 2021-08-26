package currentsong

import (
	"fmt"
	"net/http"

	"go.xsfx.dev/logginghandler"
)

var recvs = make(map[chan string]struct{}) // nolint:gochecknoglobals

// Write writes current track to the receivers.
func Write(track string) {
	for k := range recvs {
		k <- track
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	logger := logginghandler.Logger(r)
	logger.Debug().Msg("got a new receiver")

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error().Msg("streaming unsupported")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// TODO: has to be something else!
	w.Header().Set("Access-Control-Allow-Origin", "*")

	cChan := make(chan string)

	recvs[cChan] = struct{}{}

	for {
		select {
		case e := <-cChan:
			// Send event to client.
			fmt.Fprintf(w, "data: %s\n\n", e)

			// Send it right now and not buffering it.
			flusher.Flush()
		case <-r.Context().Done():
			close(cChan)
			delete(recvs, cChan)

			return
		}
	}
}
