//nolint:gochecknoglobals,golint,stylecheck
package prepare

import "embed"

//go:embed files
var Files embed.FS

//go:embed templates
var Templates embed.FS
