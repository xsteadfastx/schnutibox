//nolint:gochecknoglobals,golint,stylecheck
package assets

import (
	"embed"
)

//go:embed templates/*
//go:embed files/*
var Assets embed.FS
