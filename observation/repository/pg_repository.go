package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/observation"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type pgRepo struct {
	*pg.DB
}

func NewPgRepo(db *pg.DB) (observation.Repository, error) {
	if err := db.Model((*models.Observation)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create the 'observations' table")
	}
	return &pgRepo{db}, nil
}

func (repo *pgRepo) Store(ctx context.Context, observation *models.Observation) error {
	if _, err := repo.Model(observation).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) StoreMany(ctx context.Context, observations []*models.Observation) error {
	if _, err := repo.Model(&observations).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Update(ctx context.Context, observation *models.Observation) error {
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

func (repo *pgRepo) Fetch(ctx context.Context, f *models.ObservationFilter) ([]*models.Observation, int, error) {
	var err error
	var data []*models.Observation
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

func (repo *pgRepo) FetchServers(ctx context.Context) ([]string, error) {
	var res []string
	err := repo.
		Model(&models.Observation{}).
		Column("server").
		Context(ctx).
		Group("server").
		Order("server ASC").
		Select(&res)
	return res, err
}

func (repo *pgRepo) Delete(ctx context.Context, f *models.ObservationFilter) ([]*models.Observation, error) {
	var data []*models.Observation
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
