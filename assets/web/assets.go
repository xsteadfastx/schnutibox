//nolint:gochecknoglobals,golint,stylecheck
package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed files
var files embed.FS

var Files = sub(files, "files")

//go:embed templates
var Templates embed.FS

//go:embed swagger-ui
var swaggerUI embed.FS

var SwaggerUI = sub(swaggerUI, "swagger-ui")

func sub(f embed.FS, dir string) http.FileSystem {
	fsys, err := fs.Sub(f, dir)
	if err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("could not sub into dir")

		return nil
	}

	return http.FS(fsys)
}
