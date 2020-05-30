package models

type Server struct {
	ID                         string   `pg:",pk" json:"id" gqlgen:"id"`
	ConqueredVillagesChannelID string   `json:"conqueredVillagesChannelID" gqlgen:"conqueredVillagesChannelID"`
	LostVillagesChannelID      string   `json:"lostVillagesChannelID" gqlgen:"lostVillagesChannelID"`
	Tribes                     []*Tribe `json:"tribes,omitempty" gqlgen:"tribes"`
}

type ServerFilter struct {
	ID     []string
	Limit  int      `urlstruct:",nowhere"`
	Offset int      `urlstruct:",nowhere"`
	Order  []string `urlstruct:",nowhere"`
}
