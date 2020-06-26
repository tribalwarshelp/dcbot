package models

type Server struct {
	tableName struct{} `pg:",alias:server"`

	ID                         string       `pg:",pk" json:"id" gqlgen:"id"`
	ConqueredVillagesChannelID string       `pg:",use_zero" json:"conqueredVillagesChannelID" gqlgen:"conqueredVillagesChannelID"`
	LostVillagesChannelID      string       `pg:",use_zero" json:"lostVillagesChannelID" gqlgen:"lostVillagesChannelID"`
	Observations               Observations `json:"observation,omitempty" gqlgen:"observation"`
}

type ServerFilter struct {
	ID     []string
	Limit  int      `urlstruct:",nowhere"`
	Offset int      `urlstruct:",nowhere"`
	Order  []string `urlstruct:",nowhere"`
}
