package grpc

import (
	"context"

	fightersmodel "fightbettr.com/fighters/pkg/model"
	"fightbettr.com/gen"
	"fightbettr.com/internal/grpcutil"
	"fightbettr.com/pkg/discovery"
)

// Gateway defines an gRPC gateway for a rating service.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a rating service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// SearchFighters searches for fighters with the given status.
// It establishes a gRPC connection to the Fighters service, sends a search request,
// and returns a list of fighters.
func (g *Gateway) SearchFighters(ctx context.Context, status fightersmodel.FighterStatus) ([]*fightersmodel.Fighter, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "fighters-service", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewFightersServiceClient(conn)

	resp, err := client.SearchFighters(ctx, &gen.FightersRequest{Status: string(status)})
	if err != nil {
		return nil, err
	}

	fighters := fightersmodel.FightersFromProto(resp.Fighters)

	return fighters, nil
}
