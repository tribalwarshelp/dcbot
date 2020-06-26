package group

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
)

type Repository interface {
	Store(ctx context.Context, group *models.Group) error
	StoreMany(ctx context.Context, groups []*models.Group) error
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, filter *models.GroupFilter) ([]*models.Group, error)
	GetByID(ctx context.Context, id int) (*models.Group, error)
	Fetch(ctx context.Context, filter *models.GroupFilter) ([]*models.Group, int, error)
}
