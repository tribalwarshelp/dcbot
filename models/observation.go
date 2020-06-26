package models

type Observation struct {
	tableName struct{} `pg:",alias:observation"`

	ID       int     `json:"id" gqlgen:"id"`
	World    string  `pg:"unique:group_1,use_zero" json:"world" gqlgen:"world"`
	TribeID  int     `pg:"unique:group_1,use_zero" json:"tribeID" gqlgen:"tribeID"`
	ServerID string  `pg:"on_delete:CASCADE,unique:group_1,use_zero" json:"serverID" gqlgen:"serverID"`
	Server   *Server `json:"server,omitempty" gqlgen:"server"`
}

type Observations []*Observation

func (o Observations) Contains(world string, id int) bool {
	for _, observation := range o {
		if observation.TribeID == id && observation.World == world {
			return true
		}
	}
	return false
}

type ObservationFilter struct {
	ID       []int
	World    []string
	ServerID []string
	Limit    int      `urlstruct:",nowhere"`
	Offset   int      `urlstruct:",nowhere"`
	Order    []string `urlstruct:",nowhere"`
}
