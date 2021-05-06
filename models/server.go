package models

import (
	"github.com/Kichiyaki/gopgutil/v10"
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

func (f *ServerFilter) ApplyWithAlias(q *orm.Query, alias string) (*orm.Query, error) {
	if len(f.ID) > 0 {
		q = q.Where(gopgutil.BuildConditionArray("?"), gopgutil.AddAliasToColumnName("id", alias), pg.Array(f.ID))
	}
	return f.DefaultFilter.Apply(q)
}

func (f *ServerFilter) Apply(q *orm.Query) (*orm.Query, error) {
	return f.ApplyWithAlias(q, "server")
}
