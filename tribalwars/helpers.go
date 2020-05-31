package tribalwars

var (
	GuestEndpoints = map[string]map[string]string{
		"pl": map[string]string{
			"tribe_info":   "https://%s.plemiona.pl/guest.php?screen=info_ally&id=%d",
			"player_info":  "https://%s.plemiona.pl/guest.php?screen=info_player&id=%d",
			"village_info": "https://%s.plemiona.pl/guest.php?screen=info_village&id=%d",
		},
	}
)

func LanguageCodeFromWorldName(world string) string {
	if len(world) < 2 {
		return ""
	}
	return world[0:2]
}
