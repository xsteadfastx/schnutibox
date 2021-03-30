//nolint:gochecknoglobals,goerr113
package config

import "fmt"

var Cfg Config

type Config struct {
	// Reader is used to configure the RFID Reader.
	Reader struct {
		Dev string `mapstructure:"Dev"`
	} `mapstructure:"Reader"`

	// Box is used to configure a webinterface.
	Box struct {
		Hostname string `mapstructure:"Hostname"`
		Port     int    `mapstructure:"Port"`
	} `mapstructure:"Box"`

	// MPD contains the connection details for the Music Player Daemon.
	MPD struct {
		Hostname string
		Port     int
	} `mapstructure:"MPD"`

	// Meta contains all meta RFID's.
	Meta struct {
		Stop string `mapstructure:"Stop"`
	} `mapstructure:"Meta"`

	// Tracks contains all RFID's and its MPD URLs.
	Tracks map[string][]string `mapstructure:"Tracks"`
}

func (c *Config) Require() error {
	// RFID.
	if c.Reader.Dev == "" {
		return fmt.Errorf("missing: Reader.Dev")
	}

	// MPD.
	if c.MPD.Hostname == "" {
		return fmt.Errorf("missing: MPD.Hostname")
	}

	if c.MPD.Port == 0 {
		return fmt.Errorf("missing: MPD.Port")
	}

	// Meta.
	if c.Meta.Stop == "" {
		return fmt.Errorf("missing: Meta.Stop")
	}

	return nil
}
