package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/models"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type pgRepo struct {
	*pg.DB
}

func NewPgRepo(db *pg.DB) (group.Repository, error) {
	if err := db.Model((*models.Group)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "cannot create 'groups' table")
	}
	return &pgRepo{db}, nil
}

func (repo *pgRepo) Store(ctx context.Context, group *models.Group) error {
	if _, err := repo.Model(group).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) StoreMany(ctx context.Context, groups []*models.Group) error {
	if _, err := repo.Model(&groups).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) Update(ctx context.Context, group *models.Group) error {
	if _, err := repo.
		Model(group).
		WherePK().
		Returning("*").
		Context(ctx).
		Update(); err != nil {
		return err
	}
	return nil
}

func (repo *pgRepo) GetByID(ctx context.Context, id int) (*models.Group, error) {
	group := &models.Group{
		ID: id,
	}
	if err := repo.
		Model(group).
		WherePK().
		Returning("*").
		Relation("Observations").
		Context(ctx).
		Select(); err != nil {
		return nil, err
	}
	return group, nil
}

func (repo *pgRepo) Fetch(ctx context.Context, f *models.GroupFilter) ([]*models.Group, int, error) {
	var err error
	data := []*models.Group{}
	query := repo.Model(&data).Relation("Server").Relation("Observations").Context(ctx)

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

func (repo *pgRepo) Delete(ctx context.Context, f *models.GroupFilter) ([]*models.Group, error) {
	data := []*models.Group{}
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
