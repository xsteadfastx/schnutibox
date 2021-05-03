//nolint:gochecknoglobals
package web

import "embed"

//go:embed files
var Files embed.FS

//go:embed templates
var Templates embed.FS
