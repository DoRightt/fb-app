package psql

import (
	"context"

	"fightbettr.com/fighters/pkg/cfg"
	"fightbettr.com/fighters/pkg/logger"
	"fightbettr.com/pkg/pgxs"
)

const sep = ` AND `

// Repository represents a repository for interacting with user data in the database.
// It embeds the pgxs.Repo, which provides the basic PostgreSQL database operations.
type Repository struct {
	pgxs.FbRepo
}

// New creates and returns a new instance of User Repository using the provided logger
func New(ctx context.Context, logger logger.FbLogger) (*Repository, error) {
	db, err := pgxs.NewPool(ctx, logger, cfg.ViperPostgres())
	if err != nil {
		return nil, err
	}

	return &Repository{
		FbRepo: db,
	}, nil
}

func (r *Repository) PoolClose() {
	r.GetPool().Close()
}
