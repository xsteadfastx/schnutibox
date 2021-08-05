package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.xsfx.dev/logginghandler"
	assets "go.xsfx.dev/schnutibox/assets/web"
	"go.xsfx.dev/schnutibox/internal/config"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
	"go.xsfx.dev/schnutibox/pkg/sselog"
	"go.xsfx.dev/schnutibox/pkg/timer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

// grpcHandlerFunc reads header and returns a grpc handler or a http one.
// nolint:interfacer
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

type identifyServer struct{}

// Identify searches in tracks config for entries and returns them.
// nolint:goerr113
func (i identifyServer) Identify(ctx context.Context, in *api.IdentifyRequest) (*api.IdentifyResponse, error) {
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

type timerServer struct{}

func (t timerServer) Create(ctx context.Context, req *api.Timer) (*api.Timer, error) {
	timer.T = req

	return timer.T, nil
}

// Get just returns the status of the timer.
func (t timerServer) Get(ctx context.Context, req *api.TimerEmpty) (*api.Timer, error) {
	return timer.T, nil
}

func gw(s *grpc.Server, conn string) *runtime.ServeMux {
	ctx := context.Background()
	gopts := []grpc.DialOption{grpc.WithInsecure()}

	api.RegisterIdentifierServiceServer(s, identifyServer{})
	api.RegisterTimerServiceServer(s, timerServer{})

	// Adds reflections.
	reflection.Register(s)

	gwmux := runtime.NewServeMux()
	if err := api.RegisterIdentifierServiceHandlerFromEndpoint(ctx, gwmux, conn, gopts); err != nil {
		log.Fatal().Err(err).Msg("could not register grpc endpoint")
	}

	return gwmux
}

func Run(command *cobra.Command, args []string) {
	// Create host string for serving web.
	lh := fmt.Sprintf("%s:%d", config.Cfg.Box.Hostname, config.Cfg.Box.Port)

	// Create grpc server.
	grpcServer := grpc.NewServer()

	// Define http handlers.
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(root))

	mux.Handle("/log", http.HandlerFunc(sselog.LogHandler))

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(assets.Files)),
	)

	mux.Handle(
		"/swagger-ui/",
		http.StripPrefix("/swagger-ui/", http.FileServer(assets.SwaggerUI)),
	)

	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/api/", http.StripPrefix("/api", gw(grpcServer, lh)))

	// PPROF.
	if viper.GetBool("pprof") {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
	}

	// Serving http.
	log.Info().Msgf("serving HTTP on %s...", lh)

	log.Fatal().Err(
		http.ListenAndServe(
			lh,
			grpcHandlerFunc(
				grpcServer,
				logginghandler.Handler(mux),
			),
		),
	).Msg("goodbye")
}
