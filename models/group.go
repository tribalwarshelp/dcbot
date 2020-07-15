package models

type Group struct {
	ID                         int          `pg:",pk" json:"id" gqlgen:"id"`
	ConqueredVillagesChannelID string       `pg:",use_zero" json:"conqueredVillagesChannelID" gqlgen:"conqueredVillagesChannelID"`
	LostVillagesChannelID      string       `pg:",use_zero" json:"lostVillagesChannelID" gqlgen:"lostVillagesChannelID"`
	ServerID                   string       `pg:"on_delete:CASCADE,use_zero" json:"serverID" gqlgen:"serverID"`
	ShowEnnobledBarbarians     bool         `pg:",use_zero"`
	Server                     *Server      `json:"server,omitempty" gqlgen:"server"`
	Observations               Observations `json:"observation,omitempty" gqlgen:"observation"`
}

type GroupFilter struct {
	ID       []int
	ServerID []string
	Limit    int      `urlstruct:",nowhere"`
	Offset   int      `urlstruct:",nowhere"`
	Order    []string `urlstruct:",nowhere"`
}
