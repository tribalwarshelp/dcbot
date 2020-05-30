package tribe

import (
	"context"
	"twdcbot/models"
)

type Repository interface {
	Store(ctx context.Context, tribe *models.Tribe) error
	StoreMany(ctx context.Context, tribes []*models.Tribe) error
	Update(ctx context.Context, tribe *models.Tribe) error
	Delete(ctx context.Context, filter *models.TribeFilter) ([]*models.Tribe, error)
	Fetch(ctx context.Context, filter *models.TribeFilter) ([]*models.Tribe, int, error)
	FetchWorlds(ctx context.Context) ([]string, error)
}
