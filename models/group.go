package models

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Group struct {
	ID                         int          `pg:",pk" json:"id" gqlgen:"id"`
	ConqueredVillagesChannelID string       `pg:",use_zero" json:"conqueredVillagesChannelID" gqlgen:"conqueredVillagesChannelID"`
	LostVillagesChannelID      string       `pg:",use_zero" json:"lostVillagesChannelID" gqlgen:"lostVillagesChannelID"`
	ShowEnnobledBarbarians     bool         `pg:",use_zero"`
	ShowInternals              bool         `pg:",use_zero"`
	ServerID                   string       `pg:"on_delete:CASCADE,use_zero" json:"serverID" gqlgen:"serverID"`
	Server                     *Server      `json:"server,omitempty" gqlgen:"server" pg:"rel:has-one"`
	Observations               Observations `json:"observation,omitempty" gqlgen:"observation" pg:"rel:has-many"`
}

type GroupFilter struct {
	ID       []int
	ServerID []string
	DefaultFilter
}

func (f *GroupFilter) ApplyWithPrefix(prefix string) func(q *orm.Query) (*orm.Query, error) {
	return func(q *orm.Query) (*orm.Query, error) {
		if len(f.ID) > 0 {
			column := addPrefixToColumnName("id", prefix)
			q = q.Where(column+" = ANY(?)", pg.Array(f.ID))
		}
		if len(f.ServerID) > 0 {
			column := addPrefixToColumnName("server_id", prefix)
			q = q.Where(column+" = ANY(?)", pg.Array(f.ServerID))
		}
		return f.DefaultFilter.Apply(q)
	}
}

func (f *GroupFilter) Apply(q *orm.Query) (*orm.Query, error) {
	return f.ApplyWithPrefix("group")(q)
}
