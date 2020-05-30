package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	_cron "twdcbot/cron"
	"twdcbot/discord"
	"twdcbot/mode"
	server_repository "twdcbot/server/repository"
	tribe_repository "twdcbot/tribe/repository"

	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
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
		Status:           "Tribalwars | tw!help",
		TribeRepository:  tribeRepo,
		ServerRepository: serverRepo,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))),
	))
	_cron.AttachHandlers(c, _cron.Config{
		ServerRepo: serverRepo,
		TribeRepo:  tribeRepo,
		Discord:    sess,
	})
	go func() {
		c.Run()
	}()
	defer c.Stop()

	log.Print("Bot is waiting for your actions!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	logrus.Info("shutting down")
	os.Exit(0)

}
