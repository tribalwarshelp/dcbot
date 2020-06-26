package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type pgRepo struct {
	*pg.DB
}

func NewPgRepo(db *pg.DB) (server.Repository, error) {
	if err := db.CreateTable((*models.Server)(nil), &orm.CreateTableOptions{
		IfNotExists: true,
	}); err != nil {
		return nil, errors.Wrap(err, "Cannot create 'servers' table")
	}
	return &pgRepo{db}, nil
}

func (repo *pgRepo) Store(ctx context.Context, server *models.Server) error {
	if _, err := repo.
		Model(server).
		Where("id = ?id").
		Returning("*").
		Context(ctx).
		Relation("Observations").
		SelectOrInsert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Update(ctx context.Context, server *models.Server) error {
	if _, err := repo.
		Model(server).
		WherePK().
		Returning("*").
		Context(ctx).
		Update(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Fetch(ctx context.Context, f *models.ServerFilter) ([]*models.Server, int, error) {
	var err error
	data := []*models.Server{}
	query := repo.Model(&data).Context(ctx).Relation("Observations")

	if f != nil {
		query = query.
			WhereStruct(f).
			Limit(f.Limit).
			Offset(f.Offset)

		if len(f.Order) > 0 {
			query = query.Order(f.Order...)
		}
	}

	total, err := query.SelectAndCount()
	if err != nil && err != pg.ErrNoRows {
		return nil, 0, err
	}

	return data, total, nil
}

func (repo *pgRepo) Delete(ctx context.Context, f *models.ServerFilter) ([]*models.Server, error) {
	data := []*models.Server{}
	query := repo.Model(&data).Context(ctx)

	if f != nil {
		query = query.WhereStruct(f)
	}

	_, err := query.
		Returning("*").
		Delete()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return data, nil
}
