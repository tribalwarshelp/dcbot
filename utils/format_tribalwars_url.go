package utils

import "fmt"

const (
	endpointTribeProfile   = "/game.php?screen=info_ally&id=%d"
	endpointPlayerProfile  = "/game.php?screen=info_player&id=%d"
	endpointVillageProfile = "/game.php?screen=info_village&id=%d"
)

func FormatVillageURL(world, host string, id int) string {
	return fmt.Sprintf("https://%s.%s"+endpointVillageProfile, world, host, id)
}

func FormatTribeURL(world, host string, id int) string {
	return fmt.Sprintf("https://%s.%s"+endpointTribeProfile, world, host, id)
}

func FormatPlayerURL(world, host string, id int) string {
	return fmt.Sprintf("https://%s.%s"+endpointPlayerProfile, world, host, id)
}
