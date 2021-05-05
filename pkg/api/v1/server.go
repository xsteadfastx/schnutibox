package v1

import (
	"context"

	"github.com/rs/zerolog/log"
)

type Server struct {
	UnimplementedIdentifierServer
}

func (s Server) Identify(ctx context.Context, r *IdentifyRequest) (*Tracks, error) {
	log.Info().Msg("Tracks und so")

	return &Tracks{}, nil
}
