package models

type Tribe struct {
	ID       int     `json:"id" gqlgen:"id"`
	World    string  `pg:"unique:group_1" json:"world" gqlgen:"world"`
	TribeID  int     `pg:"unique:group_1" json:"TribeID" gqlgen:"TribeID"`
	ServerID string  `pg:"on_delete:CASCADE,unique:group_1" json:"serverID" gqlgen:"serverID"`
	Server   *Server `json:"server,omitempty" gqlgen:"server"`
}

type TribeFilter struct {
	ID     []string
	Limit  int      `urlstruct:",nowhere"`
	Offset int      `urlstruct:",nowhere"`
	Order  []string `urlstruct:",nowhere"`
}
