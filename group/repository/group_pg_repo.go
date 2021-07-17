package repository

import (
	"context"

	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/model"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

type PGRepository struct {
	*pg.DB
}

var _ group.Repository = &PGRepository{}

func NewPgRepo(db *pg.DB) (*PGRepository, error) {
	if err := db.Model((*model.Group)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists:   true,
		FKConstraints: true,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't create the 'groups' table")
	}
	return &PGRepository{db}, nil
}

func (repo *PGRepository) Store(ctx context.Context, group *model.Group) error {
	if _, err := repo.Model(group).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) StoreMany(ctx context.Context, groups []*model.Group) error {
	if _, err := repo.Model(&groups).Returning("*").Context(ctx).Insert(); err != nil {
		return err
	}
	return nil
}

func (repo *PGRepository) Update(ctx context.Context, group *model.Group) error {
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

func (repo *PGRepository) GetByID(ctx context.Context, id int) (*model.Group, error) {
	g := &model.Group{
		ID: id,
	}
	if err := repo.
		Model(g).
		WherePK().
		Returning("*").
		Relation("Observations").
		Context(ctx).
		Select(); err != nil {
		return nil, err
	}
	return g, nil
}

func (repo *PGRepository) Fetch(ctx context.Context, f *model.GroupFilter) ([]*model.Group, int, error) {
	var err error
	var data []*model.Group
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

func (repo *PGRepository) Delete(ctx context.Context, f *model.GroupFilter) ([]*model.Group, error) {
	var data []*model.Group
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
