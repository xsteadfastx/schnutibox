package web

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.xsfx.dev/logginghandler"
	assets "go.xsfx.dev/schnutibox/assets/web"
	"go.xsfx.dev/schnutibox/internal/config"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
	"go.xsfx.dev/schnutibox/pkg/sselog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

type server struct{}

// Identify searches in tracks config for entries and returns them.
// nolint:goerr113
func (s server) Identify(ctx context.Context, in *api.IdentifyRequest) (*api.IdentifyResponse, error) {
	r := &api.IdentifyResponse{}

	if in.Id == "" {
		return r, fmt.Errorf("no id in request specified")
	}

	t, ok := config.Cfg.Tracks[in.Id]
	if !ok {
		return r, fmt.Errorf("could not find track for id: %s", in.Id)
	}

	r.Name = t.Name
	r.Uris = t.Uris

	return r, nil
}

func gw(conn string) *runtime.ServeMux {
	ctx := context.Background()
	gopts := []grpc.DialOption{grpc.WithInsecure()}
	s := grpc.NewServer()

	api.RegisterIdentifierServiceServer(s, server{})
	reflection.Register(s)

	lis, err := net.Listen("tcp", conn)
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen")
	}

	log.Info().Msgf("serving GRPC on %s...", conn)

	go func() {
		log.Fatal().Err(s.Serve(lis))
	}()

	gwmux := runtime.NewServeMux()
	if err := api.RegisterIdentifierServiceHandlerFromEndpoint(ctx, gwmux, conn, gopts); err != nil {
		log.Fatal().Err(err).Msg("could not register grpc endpoint")
	}

	return gwmux
}

func Run(command *cobra.Command, args []string) {
	// Create host string for serving web.
	lh := fmt.Sprintf("%s:%d", config.Cfg.Box.Hostname, config.Cfg.Box.Port)
	lg := fmt.Sprintf("%s:%d", config.Cfg.Box.Hostname, config.Cfg.Box.Grpc)

	// Define http handlers.
	mux := http.NewServeMux()
	mux.Handle("/", logginghandler.Handler(http.HandlerFunc(root)))
	mux.Handle("/log", logginghandler.Handler(http.HandlerFunc(sselog.LogHandler)))
	mux.Handle(
		"/static/",
		logginghandler.Handler(
			http.StripPrefix("/static/", http.FileServer(http.FS(assets.Files))),
		),
	)
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/api/", http.StripPrefix("/api", gw(lg)))

	// Serving http.
	log.Info().Msgf("serving HTTP on %s...", lh)

	log.Fatal().Err(http.ListenAndServe(lh, logginghandler.Handler(mux))).Msg("goodbye")
}
