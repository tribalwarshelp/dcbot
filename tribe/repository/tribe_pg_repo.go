package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/tribe"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type pgRepo struct {
	*pg.DB
}

func NewPgRepo(db *pg.DB) (tribe.Repository, error) {
	if err := db.CreateTable((*models.Tribe)(nil), &orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "Cannot create 'tribes' table")
	}
	return &pgRepo{db}, nil
}

func (repo *pgRepo) Store(ctx context.Context, tribe *models.Tribe) error {
	if _, err := repo.Model(tribe).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) StoreMany(ctx context.Context, tribes []*models.Tribe) error {
	if _, err := repo.Model(&tribes).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Update(ctx context.Context, tribe *models.Tribe) error {
	if _, err := repo.
		Model(tribe).
		WherePK().
		Returning("*").
		Context(ctx).
		UpdateNotZero(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Fetch(ctx context.Context, f *models.TribeFilter) ([]*models.Tribe, int, error) {
	var err error
	data := []*models.Tribe{}
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
	data := []*models.Tribe{}
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

func (repo *pgRepo) Delete(ctx context.Context, f *models.TribeFilter) ([]*models.Tribe, error) {
	data := []*models.Tribe{}
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
