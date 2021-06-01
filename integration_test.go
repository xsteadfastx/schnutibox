// +build integration

package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	sshclient "github.com/helloyi/go-sshclient"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tmc/scp"
	"go.xsfx.dev/don"
	"golang.org/x/crypto/ssh"
)

const (
	sdcard           = "/SWAP/2021-05-07-raspios-buster-armhf-lite.img"
	sshUser          = "pi"
	sshPass          = "raspberry"
	sshHost          = "docker"
	containerTimeout = 15 * time.Minute
)

// Variables used for accessing stuff in the test functions.
// nolint:gochecknoglobals
var (
	sshConn string
	client  *sshclient.Client
)

// raspbianWorkCopy creates a temp image file.
func raspbianWorkCopy() (string, error) {
	f, err := ioutil.TempFile("", "schnutibox")
	if err != nil {
		return "", fmt.Errorf("could not create tempfile: %w", err)
	}

	o, err := os.Open(sdcard)
	if err != nil {
		return "", fmt.Errorf("could not open sdcard image: %w", err)
	}

	defer o.Close()

	if _, err := io.Copy(f, o); err != nil {
		return "", fmt.Errorf("could not copy image: %w", err)
	}

	return f.Name(), nil
}

func copySchnutibox(user, pass, host string) error {
	// Local file.
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get pwd: %w", err)
	}

	binPath := pwd + "/dist/schnutibox_linux_arm_6/schnutibox"

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return fmt.Errorf("could not create ssh client: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("could not create session: %w", err)
	}

	log.Debug().Str("schnutibox path", binPath).Msg("copy binary")

	if err := scp.CopyPath(binPath, "/tmp/schnutibox", session); err != nil {
		return fmt.Errorf("could not scp binary: %w", err)
	}

	return nil
}

// teardown removes some temp test stuff.
func teardown(pool *dockertest.Pool, resource *dockertest.Resource) {
	log.Info().Msg("getting rid of container")

	if err := pool.Purge(resource); err != nil {
		log.Fatal().Err(err).Msg("could not cleanup")
	}
}

// nolint:funlen
func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to docker")
	}

	if err != nil {
		log.Fatal().Err(err).Msg("could not get pwd")
	}

	log.Info().Msg("starting container")

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository:   "lukechilds/dockerpi",
			Tag:          "vm",
			Mounts:       []string{sdcard + ":/sdcard/filesystem.img"},
			ExposedPorts: []string{"5022/tcp"},
			PortBindings: map[dc.Port][]dc.PortBinding{
				"5022/tcp": {{HostIP: "0.0.0.0", HostPort: "5022"}},
			},
		})
	if err != nil {
		log.Fatal().Err(err).Msg("could not start resource")
	}

	// Starting container.
	log.Info().Msg("waiting to be ready")

	sshConn = sshHost + ":" + resource.GetPort("5022/tcp")

	// Readiness.
	if err := don.Check(
		func() bool {
			client, err := sshclient.DialWithPasswd(sshConn, sshUser, sshPass)
			if err == nil {
				if err := client.Close(); err != nil {
					return false
				}

				return true
			}

			return false
		},
		20*time.Minute,
		15*time.Second,
	); err != nil {
		log.Error().Err(err).Msg("timeout. could not connect to container via ssh")
		teardown(pool, resource)
		os.Exit(1)
	}

	// Connect via SSH.
	log.Info().Msg("connect via ssh")

	client, err = sshclient.DialWithPasswd(sshConn, sshUser, sshPass)
	if err != nil {
		log.Error().Err(err).Msg("could not connect via ssh")
		teardown(pool, resource)
		os.Exit(1)
	}

	// Copy schnutibox binary to container.
	log.Info().
		Str("user", sshUser).
		Str("pass", sshPass).
		Str("conn", sshConn).
		Msg("copy schnutibox")

	if err := copySchnutibox(sshUser, sshPass, sshConn); err != nil {
		log.Error().Err(err).Msg("could not copy schnutibox")
		teardown(pool, resource)
		os.Exit(1)
	}

	// Move schnutibox to /usr/local/bin
	log.Info().Msg("move binary to /usr/local/bin")

	if err := client.
		Cmd("sudo mv /tmp/schnutibox /usr/local/bin/schnutibox").
		SetStdio(os.Stdout, os.Stderr).
		Run(); err != nil {
		log.Error().Err(err).Msg("could not create /usr/local/bin on container")
		teardown(pool, resource)
		os.Exit(1)
	}

	// Doing the testing.
	log.Info().Msg("doing to testing")

	// Running the tests.
	code := m.Run()

	// Removing container.
	teardown(pool, resource)

	os.Exit(code)
}

// nolint:paralleltest
func TestPrepare(t *testing.T) {
	if err := client.
		Cmd("sudo schnutibox prepare --read-only").
		SetStdio(os.Stdout, os.Stderr).
		Run(); err != nil {
		log.Error().Err(err).Msg("could not run command")
		t.Fatal()
	}
}

// nolint:paralleltest
func TestBoxService(t *testing.T) {
	if err := client.
		Cmd("file /etc/systemd/system/schnutibox.service").
		SetStdio(os.Stdout, os.Stderr).
		Run(); err != nil {
		log.Error().Err(err).Msg("could not find schnutibox service file")
		t.Fatal()
	}
}

// nolint:paralleltest
func TestUdevRules(t *testing.T) {
	if err := client.
		Cmd("file /etc/udev/rules.d/50-neuftech.rules").
		SetStdio(os.Stdout, os.Stderr).
		Run(); err != nil {
		log.Error().Err(err).Msg("could not find udev rules file")
		t.Fatal()
	}
}
