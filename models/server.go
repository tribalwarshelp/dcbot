package models

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Server struct {
	tableName struct{} `pg:",alias:server"`

	ID                string   `pg:",pk" json:"id" gqlgen:"id"`
	Lang              string   `pg:",use_zero"`
	CoordsTranslation string   `pg:",use_zero"`
	Groups            []*Group `pg:"rel:has-many"`
}

type ServerFilter struct {
	ID []string
	DefaultFilter
}

func (f *ServerFilter) ApplyWithPrefix(prefix string) func(q *orm.Query) (*orm.Query, error) {
	return func(q *orm.Query) (*orm.Query, error) {
		if len(f.ID) > 0 {
			column := addPrefixToColumnName("id", prefix)
			q = q.Where(column+" = ANY(?)", pg.Array(f.ID))
		}
		return f.DefaultFilter.Apply(q)
	}
}

func (f *ServerFilter) Apply(q *orm.Query) (*orm.Query, error) {
	return f.ApplyWithPrefix("server")(q)
}
