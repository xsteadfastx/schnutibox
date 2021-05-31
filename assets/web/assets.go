//nolint:gochecknoglobals
package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed files
var files embed.FS

// Files is the sub directed http.FileSystem for files.
var Files = sub(files, "files")

// Templates stores the templates.
//go:embed templates
var Templates embed.FS

//go:embed swagger-ui
var swaggerUI embed.FS

// SwaggerUI is the sub directed http.FileSystem for the swagger-ui.
var SwaggerUI = sub(swaggerUI, "swagger-ui")

func sub(f embed.FS, dir string) http.FileSystem {
	fsys, err := fs.Sub(f, dir)
	if err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("could not sub into dir")

		return nil
	}

	return http.FS(fsys)
}
