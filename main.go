package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"twdcbot/discord"
	"twdcbot/mode"
	server_repository "twdcbot/server/repository"
	tribe_repository "twdcbot/tribe/repository"

	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func init() {
	os.Setenv("TZ", "UTC")

	if mode.Get() == mode.DevelopmentMode {
		godotenv.Load(".env.development")
	}
}

func main() {
	//postgres
	db := pg.Connect(&pg.Options{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		Addr:     os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
	})
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Fatal(err)
		}
	}()
	serverRepo, err := server_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	tribeRepo, err := tribe_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	sess, err := discord.New(discord.SessionConfig{
		Token:            os.Getenv("BOT_TOKEN"),
		CommandPrefix:    "tw!",
		Status:           "Twstats",
		TribeRepository:  tribeRepo,
		ServerRepository: serverRepo,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	// for world, conquers := range scraper.New([]string{"pl149", "pl150"}, time.Now().Add(time.Minute*-10)).Scrap() {
	// 	fmt.Print("\n\n", world, "\n\n")
	// 	for _, c := range conquers {
	// 		log.Print(c.ConqueredAt,
	// 			" | ",
	// 			c.VillageID,
	// 			" | ",
	// 			c.Village,
	// 			" | ",
	// 			c.OldOwnerID,
	// 			" | ",
	// 			c.OldOwnerName,
	// 			" | ",
	// 			c.OldOwnerTribeID,
	// 			" | ",
	// 			c.OldOwnerTribeName,
	// 			" | ",
	// 			c.NewOwnerID,
	// 			" | ",
	// 			c.NewOwnerName,
	// 			" | ",
	// 			c.NewOwnerTribeID,
	// 			" | ",
	// 			c.NewOwnerTribeName)
	// 	}
	// }

	log.Print("Bot is waiting for your actions!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	logrus.Info("shutting down")
	os.Exit(0)

}
