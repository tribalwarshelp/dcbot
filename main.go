package main

import (
	"fmt"
	"log"
	"time"
	"twdcbot/scraper"
)

func main() {
	for world, conquers := range scraper.New([]string{"pl149", "pl150"}, time.Now().Add(time.Minute*-10)).Scrap() {
		fmt.Print("\n\n", world, "\n\n")
		for _, c := range conquers {
			log.Print(c.ConqueredAt,
				" | ",
				c.VillageID,
				" | ",
				c.Village,
				" | ",
				c.OldOwnerID,
				" | ",
				c.OldOwnerName,
				" | ",
				c.OldOwnerTribeID,
				" | ",
				c.OldOwnerTribeName,
				" | ",
				c.NewOwnerID,
				" | ",
				c.NewOwnerName,
				" | ",
				c.NewOwnerTribeID,
				" | ",
				c.NewOwnerTribeName)
		}
	}
}
