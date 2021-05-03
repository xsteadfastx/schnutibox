// package sselog is work in progress to implement a writer that sends its logs
// to a http server side event.
// nolint:gochecknoglobals,godox
package sselog

import (
	"fmt"
	"net/http"

	"go.xsfx.dev/logginghandler"
)

var Log *SSELog

type SSELog struct {
	Receivers map[chan []byte]struct{}
}

func NewSSELog() *SSELog {
	return &SSELog{
		Receivers: make(map[chan []byte]struct{}),
	}
}

func (l SSELog) Write(p []byte) (n int, err error) {
	// Send log message to all receiver channels.
	for r := range l.Receivers {
		r <- p
	}

	return len(p), nil
}

func LogHandler(w http.ResponseWriter, r *http.Request) {
	logger := logginghandler.Logger(r)

	logger.Info().Msg("registering a new sse logger")

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

	cChan := make(chan []byte)

	Log.Receivers[cChan] = struct{}{}

	for {
		select {
		case e := <-cChan:
			// Send event to client.
			fmt.Fprintf(w, "data: %s\n\n", e)

			// Send it right now and not buffering it.
			flusher.Flush()
		case <-r.Context().Done():
			close(cChan)
			delete(Log.Receivers, cChan)

			return
		}
	}
}
