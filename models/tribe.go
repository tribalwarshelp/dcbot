package models

type Tribe struct {
	ID       int     `json:"id" gqlgen:"id"`
	World    string  `pg:"unique:group_1" json:"world" gqlgen:"world"`
	TribeID  int     `pg:"unique:group_1" json:"TribeID" gqlgen:"TribeID"`
	ServerID string  `pg:"on_delete:CASCADE,unique:group_1" json:"serverID" gqlgen:"serverID"`
	Server   *Server `json:"server,omitempty" gqlgen:"server"`
}

type Tribes []*Tribe

func (t Tribes) Contains(world string, id int) bool {
	for _, tribe := range t {
		if tribe.TribeID == id && tribe.World == world {
			return true
		}
	}
	return false
}

type TribeFilter struct {
	ID       []int
	ServerID []string
	Limit    int      `urlstruct:",nowhere"`
	Offset   int      `urlstruct:",nowhere"`
	Order    []string `urlstruct:",nowhere"`
}
