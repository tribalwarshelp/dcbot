package models

type Group struct {
	ID                         int          `pg:",pk" json:"id" gqlgen:"id"`
	ConqueredVillagesChannelID string       `pg:",use_zero" json:"conqueredVillagesChannelID" gqlgen:"conqueredVillagesChannelID"`
	LostVillagesChannelID      string       `pg:",use_zero" json:"lostVillagesChannelID" gqlgen:"lostVillagesChannelID"`
	ShowEnnobledBarbarians     bool         `pg:",use_zero"`
	ServerID                   string       `pg:"on_delete:CASCADE,use_zero" json:"serverID" gqlgen:"serverID"`
	Server                     *Server      `json:"server,omitempty" gqlgen:"server"`
	Observations               Observations `json:"observation,omitempty" gqlgen:"observation"`
}

type GroupFilter struct {
	tableName struct{} `urlstruct:"group"`

	ID       []int
	ServerID []string
	Limit    int      `urlstruct:",nowhere"`
	Offset   int      `urlstruct:",nowhere"`
	Order    []string `urlstruct:",nowhere"`
}
