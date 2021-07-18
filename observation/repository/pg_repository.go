package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/observation"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type PGRepository struct {
	*pg.DB
}

var _ observation.Repository = &PGRepository{}

func NewPgRepository(db *pg.DB) (*PGRepository, error) {
	if err := db.Model((*model.Observation)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create the 'observations' table")
	}
	return &PGRepository{db}, nil
}

func (repo *PGRepository) Store(ctx context.Context, observation *model.Observation) error {
	if _, err := repo.Model(observation).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) StoreMany(ctx context.Context, observations []*model.Observation) error {
	if _, err := repo.Model(&observations).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) Update(ctx context.Context, observation *model.Observation) error {
	if _, err := repo.
		Model(observation).
		WherePK().
		Returning("*").
		Context(ctx).
		UpdateNotZero(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) Fetch(ctx context.Context, f *model.ObservationFilter) ([]*model.Observation, int, error) {
	var err error
	var data []*model.Observation
	query := repo.Model(&data).Context(ctx)

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

func (repo *PGRepository) FetchServers(ctx context.Context) ([]string, error) {
	var res []string
	err := repo.
		Model(&model.Observation{}).
		Column("server").
		Context(ctx).
		Group("server").
		Order("server ASC").
		Select(&res)
	return res, err
}

func (repo *PGRepository) Delete(ctx context.Context, f *model.ObservationFilter) ([]*model.Observation, error) {
	var data []*model.Observation
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
