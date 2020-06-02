package server

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
)

type Repository interface {
	Store(ctx context.Context, server *models.Server) error
	Update(ctx context.Context, server *models.Server) error
	Delete(ctx context.Context, filter *models.ServerFilter) ([]*models.Server, error)
	Fetch(ctx context.Context, filter *models.ServerFilter) ([]*models.Server, int, error)
}
