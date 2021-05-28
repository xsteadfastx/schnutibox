//nolint:gochecknoglobals
package prepare

import "embed"

// Files are files to be copied to the system.
//go:embed files
var Files embed.FS

// Templates are the used templates for creating file on the system.
//go:embed templates
var Templates embed.FS
