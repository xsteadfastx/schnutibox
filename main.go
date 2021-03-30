// thanks to: https://medium.com/coinmonks/iot-tutorial-read-tags-from-a-usb-rfid-reader-with-raspberry-pi-and-node-red-from-scratch-4554836be127
//nolint:lll,godox
package main

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.xsfx.dev/schnutibox/cmd"
)

func main() {
	// TODO: Using io.MultiWriter here to implement a SSE Logger at some point.
	log.Logger = zerolog.New(io.MultiWriter(os.Stderr)).With().Caller().Logger()

	cmd.Execute()
}
