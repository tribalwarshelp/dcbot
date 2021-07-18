package observation

import (
	"context"

	"github.com/tribalwarshelp/dcbot/model"
)

type Repository interface {
	Store(ctx context.Context, observation *model.Observation) error
	StoreMany(ctx context.Context, observations []*model.Observation) error
	Update(ctx context.Context, observation *model.Observation) error
	Delete(ctx context.Context, filter *model.ObservationFilter) ([]*model.Observation, error)
	Fetch(ctx context.Context, filter *model.ObservationFilter) ([]*model.Observation, int, error)
	FetchServers(ctx context.Context) ([]string, error)
}
