package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type PGRepository struct {
	*pg.DB
}

var _ server.Repository = &PGRepository{}

func NewPgRepository(db *pg.DB) (*PGRepository, error) {
	if err := db.Model((*model.Server)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create the 'servers' table")
	}
	return &PGRepository{db}, nil
}

func (repo *PGRepository) Store(ctx context.Context, server *model.Server) error {
	if _, err := repo.
		Model(server).
		Where("id = ?id").
		Returning("*").
		Context(ctx).
		Relation("Groups").
		SelectOrInsert(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) Update(ctx context.Context, server *model.Server) error {
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

func (repo *PGRepository) Fetch(ctx context.Context, f *model.ServerFilter) ([]*model.Server, int, error) {
	var err error
	var data []*model.Server
	query := repo.Model(&data).Context(ctx).Relation("Groups")

	if f != nil {
		query = query.
			Apply(f.Apply).
			Limit(f.Limit).
			Offset(f.Offset)
	}

	total, err := query.SelectAndCount()
	if err != nil && err != pg.ErrNoRows {
		return nil, 0, err
	}

	return data, total, nil
}

func (repo *PGRepository) Delete(ctx context.Context, f *model.ServerFilter) ([]*model.Server, error) {
	var data []*model.Server
	query := repo.Model(&data).Context(ctx)

	if f != nil {
		query = query.Apply(f.Apply)
	}

	_, err := query.
		Returning("*").
		Delete()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return data, nil
}
