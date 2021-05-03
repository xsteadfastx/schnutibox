//nolint:gochecknoglobals,golint,stylecheck
package web

import "embed"

//go:embed files
var Files embed.FS

//go:embed templates
var Templates embed.FS
