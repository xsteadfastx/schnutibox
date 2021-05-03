package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/logginghandler"
	assets "go.xsfx.dev/schnutibox/assets/web"
	"go.xsfx.dev/schnutibox/internal/config"
	"go.xsfx.dev/schnutibox/pkg/sselog"
)

func root(w http.ResponseWriter, r *http.Request) {
	logger := logginghandler.Logger(r)

	t, err := template.ParseFS(assets.Templates, "templates/index.html.tmpl")
	if err != nil {
		logger.Error().Err(err).Msg("could not parse template")
		http.Error(w, "could not parse template", http.StatusInternalServerError)

		return
	}

	if err := t.Execute(w, struct{}{}); err != nil {
		logger.Error().Err(err).Msg("could not execute template")
		http.Error(w, "could not execute template", http.StatusInternalServerError)

		return
	}
}

func Run(command *cobra.Command, args []string) {
	// Create host string for serving web.
	l := fmt.Sprintf("%s:%d", config.Cfg.Box.Hostname, config.Cfg.Box.Port)

	// Define http handlers.
	http.Handle("/", logginghandler.Handler(http.HandlerFunc(root)))
	http.Handle("/log", logginghandler.Handler(http.HandlerFunc(sselog.LogHandler)))
	http.Handle(
		"/static/",
		logginghandler.Handler(
			http.StripPrefix("/static/", http.FileServer(http.FS(assets.Files))),
		),
	)
	http.Handle("/metrics", promhttp.Handler())

	// Serving this thing.
	log.Info().Msgf("serving on %s...", l)

	log.Fatal().Err(http.ListenAndServe(l, nil)).Msg("goodbye")
}
