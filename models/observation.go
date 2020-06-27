package models

import "time"

type Observation struct {
	tableName struct{} `pg:",alias:observation"`

	ID        int       `json:"id" gqlgen:"id"`
	Server    string    `pg:"unique:group_1,use_zero" json:"server" gqlgen:"server"`
	TribeID   int       `pg:"unique:group_1,use_zero" json:"tribeID" gqlgen:"tribeID"`
	GroupID   int       `pg:"on_delete:CASCADE,unique:group_1,use_zero" json:"groupID" gqlgen:"groupID"`
	Group     *Group    `json:"group,omitempty" gqlgen:"group"`
	CreatedAt time.Time `pg:"default:now()" json:"createdAt" gqlgen:"createdAt" xml:"createdAt"`
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
	Limit   int      `urlstruct:",nowhere"`
	Offset  int      `urlstruct:",nowhere"`
	Order   []string `urlstruct:",nowhere"`
}
