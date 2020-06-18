package observation

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
)

type Repository interface {
	Store(ctx context.Context, observation *models.Observation) error
	StoreMany(ctx context.Context, observations []*models.Observation) error
	Update(ctx context.Context, observation *models.Observation) error
	Delete(ctx context.Context, filter *models.ObservationFilter) ([]*models.Observation, error)
	Fetch(ctx context.Context, filter *models.ObservationFilter) ([]*models.Observation, int, error)
	FetchWorlds(ctx context.Context) ([]string, error)
}
