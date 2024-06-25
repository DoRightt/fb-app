package event

import (
	"context"
	"errors"

	lg "fightbettr.com/events/pkg/logger"
	eventmodel "fightbettr.com/events/pkg/model"
	"fightbettr.com/pkg/pgxs"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type eventRepository interface {
	pgxs.FbRepo

	TxCreateEvent(ctx context.Context, tx pgx.Tx, e *eventmodel.EventRequest) (int32, error)
	TxCreateEventFight(ctx context.Context, tx pgx.Tx, f eventmodel.Fight) error
	SearchEventsCount(ctx context.Context) (int32, error)
	SearchEvents(ctx context.Context) ([]*eventmodel.Event, error)
	TxCreateBet(ctx context.Context, tx pgx.Tx, req *eventmodel.Bet) (int32, error)
	SearchBetsCount(ctx context.Context, userId int32) (int32, error)
	SearchBets(ctx context.Context, userId int32) ([]*eventmodel.Bet, error)
	SetFightResult(ctx context.Context, tx pgx.Tx, fr *eventmodel.FightResultRequest) error
	GetEventId(ctx context.Context, tx pgx.Tx, fightId int32) (int32, error)
	GetUndoneFightsCount(ctx context.Context, tx pgx.Tx, eventId int32) (int, error)
	SetEventDone(ctx context.Context, tx pgx.Tx, eventId int32) error
}

// Controller defines a metadata service controller.
type Controller struct {
	repo   eventRepository
	Logger lg.FbLogger
}

// New creates a Event service controller.
func New(repo eventRepository) *Controller {
	return &Controller{
		repo:   repo,
		Logger: lg.GetSugared(),
	}
}

// GracefulShutdown initiates a graceful shutdown of the controller,
// logging the received signal and shutting down the associated repository if available.
func (c *Controller) GracefulShutdown(ctx context.Context, sig string) {
	c.Logger.Warnf("Graceful shutdown. Signal received: %s", sig)
	if c.repo != nil {
		c.repo.GracefulShutdown()
	}
}
