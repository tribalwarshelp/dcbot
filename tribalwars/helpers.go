package tribalwars

func LanguageCodeFromWorldName(world string) string {
	if len(world) < 2 {
		return ""
	}
	return world[0:2]
}
