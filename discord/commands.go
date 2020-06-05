package discord

type Command string

var (
	HelpCommand              Command = "help"
	AddCommand               Command = "add"
	ListCommand              Command = "list"
	DeleteCommand            Command = "delete"
	LostVillagesCommand      Command = "lostvillages"
	ConqueredVillagesCommand Command = "conqueredvillages"
	TopAttCommand            Command = "topatt"
	TopDefCommand            Command = "topdef"
	TopSuppCommand           Command = "topsupp"
	TopTotalCommand          Command = "toptotal"
	TopPointsCommand         Command = "toppoints"
)

func (cmd Command) String() string {
	return string(cmd)
}

func (cmd Command) WithPrefix(prefix string) string {
	return prefix + cmd.String()
}
