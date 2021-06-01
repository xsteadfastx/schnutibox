// Package don is a little helper if you need to check for the readiness of something.
// This could be a command to run (like ssh) or a `db.Ping()` for check of the readiness
// of a database container.
package don

import (
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

var errTimeout = errors.New("timeout")

// Cmd returns a `func() bool` for working with don.Check. It executes a command and
// returns a true if everything looks fine or a false if there was some kind of error.
func Cmd(c string) func() bool {
	return func() bool {
		cmd := exec.Command("sh", "-c", c)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Warn().Err(err).Msg("cmd has error")

			return false
		}

		return true
	}
}

// Check takes a function that executes something and returns a bool to indicate if
// something is ready or not. It returns an error if it timeouts.
func Check(f func() bool, timeout time.Duration, retry time.Duration) error {
	chReady := make(chan struct{})

	go func() {
		for {
			if f() {
				chReady <- struct{}{}

				return
			}

			<-time.After(retry)
			log.Info().Msg("retrying")
		}
	}()

	select {
	case <-chReady:
		return nil

	case <-time.After(timeout):
		return errTimeout
	}
}
