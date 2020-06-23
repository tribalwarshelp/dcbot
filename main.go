package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	_cron "github.com/tribalwarshelp/dcbot/cron"
	"github.com/tribalwarshelp/dcbot/discord"
	observation_repository "github.com/tribalwarshelp/dcbot/observation/repository"
	server_repository "github.com/tribalwarshelp/dcbot/server/repository"

	"github.com/tribalwarshelp/shared/mode"

	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func init() {
	os.Setenv("TZ", "UTC")

	if mode.Get() == mode.DevelopmentMode {
		godotenv.Load(".env.development")
	}
}

func main() {
	api := sdk.New(os.Getenv("API_URL"))
	//postgres
	db := pg.Connect(&pg.Options{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		Addr:     os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
	})
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	serverRepo, err := server_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	observationRepo, err := observation_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	sess, err := discord.New(discord.SessionConfig{
		Token:                 os.Getenv("BOT_TOKEN"),
		CommandPrefix:         "tw!",
		Status:                "Tribal Wars | tw!help",
		ObservationRepository: observationRepo,
		ServerRepository:      serverRepo,
		API:                   api,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))),
	))
	_cron.Attach(c, _cron.Config{
		ServerRepo:      serverRepo,
		ObservationRepo: observationRepo,
		Discord:         sess,
		API:             api,
	})
	go func() {
		c.Run()
	}()
	defer c.Stop()

	log.Print("Bot is waiting for your actions!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	log.Print("shutting down")
}
