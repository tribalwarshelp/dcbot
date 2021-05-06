package models

import (
	"github.com/Kichiyaki/gopgutil/v10"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Observation struct {
	tableName struct{} `pg:",alias:observation"`

	ID        int            `json:"id" gqlgen:"id"`
	Server    string         `pg:"unique:group_1,use_zero" json:"server" gqlgen:"server"`
	TribeID   int            `pg:"unique:group_1,use_zero" json:"tribeID" gqlgen:"tribeID"`
	Tribe     *twmodel.Tribe `pg:"-"`
	GroupID   int            `pg:"on_delete:CASCADE,unique:group_1,use_zero" json:"groupID" gqlgen:"groupID"`
	Group     *Group         `json:"group,omitempty" gqlgen:"group" pg:"rel:has-one"`
	CreatedAt time.Time      `pg:"default:now()" json:"createdAt" gqlgen:"createdAt" xml:"createdAt"`
}

type Observations []*Observation

func (o Observations) Contains(server string, id int) bool {
	for _, observation := range o {
		if observation.TribeID == id && observation.Server == server {
			return true
		}
	}
	return false
}

type ObservationFilter struct {
	ID      []int
	Server  []string
	GroupID []int
	DefaultFilter
}

func (f *ObservationFilter) ApplyWithAlias(q *orm.Query, alias string) (*orm.Query, error) {
	if len(f.ID) > 0 {
		q = q.Where(gopgutil.BuildConditionArray("?"), gopgutil.AddAliasToColumnName("id", alias), pg.Array(f.ID))
	}
	if len(f.Server) > 0 {
		q = q.Where(gopgutil.BuildConditionArray("?"), gopgutil.AddAliasToColumnName("server", alias), pg.Array(f.Server))
	}
	if len(f.GroupID) > 0 {
		q = q.Where(gopgutil.BuildConditionArray("?"), gopgutil.AddAliasToColumnName("group_id", alias), pg.Array(f.GroupID))
	}
	return f.DefaultFilter.Apply(q)
}

func (f *ObservationFilter) Apply(q *orm.Query) (*orm.Query, error) {
	return f.ApplyWithAlias(q, "observation")
}
