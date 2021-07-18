package server

import (
	"context"

	"github.com/tribalwarshelp/dcbot/model"
)

type Repository interface {
	Store(ctx context.Context, server *model.Server) error
	Update(ctx context.Context, server *model.Server) error
	Delete(ctx context.Context, filter *model.ServerFilter) ([]*model.Server, error)
	Fetch(ctx context.Context, filter *model.ServerFilter) ([]*model.Server, int, error)
}
