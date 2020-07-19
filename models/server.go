package models

type Server struct {
	tableName struct{} `pg:",alias:server"`

	ID     string `pg:",pk" json:"id" gqlgen:"id"`
	Lang   string `pg:",use_zero"`
	Groups []*Group
}
type ServerFilter struct {
	ID     []string
	Limit  int      `urlstruct:",nowhere"`
	Offset int      `urlstruct:",nowhere"`
	Order  []string `urlstruct:",nowhere"`
}
