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
	if err := db.CreateTable((*models.Observation)(nil), &orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "Cannot create 'observations' table")
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
	data := []*models.Observation{}
	query := repo.Model(&data).Context(ctx)

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

func (repo *pgRepo) FetchWorlds(ctx context.Context) ([]string, error) {
	data := []*models.Observation{}
	res := []string{}
	err := repo.
		Model(&data).
		Column("world").
		Context(ctx).
		Group("world").
		Order("world ASC").
		Select(&res)
	return res, err
}

func (repo *pgRepo) Delete(ctx context.Context, f *models.ObservationFilter) ([]*models.Observation, error) {
	data := []*models.Observation{}
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
