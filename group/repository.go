package group

import (
	"context"

	"github.com/tribalwarshelp/dcbot/model"
)

type Repository interface {
	Store(ctx context.Context, group *model.Group) error
	StoreMany(ctx context.Context, groups []*model.Group) error
	Update(ctx context.Context, group *model.Group) error
	Delete(ctx context.Context, filter *model.GroupFilter) ([]*model.Group, error)
	GetByID(ctx context.Context, id int) (*model.Group, error)
	Fetch(ctx context.Context, filter *model.GroupFilter) ([]*model.Group, int, error)
}
