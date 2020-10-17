package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	_cron "github.com/tribalwarshelp/dcbot/cron"
	"github.com/tribalwarshelp/dcbot/discord"
	group_repository "github.com/tribalwarshelp/dcbot/group/repository"
	observation_repository "github.com/tribalwarshelp/dcbot/observation/repository"
	server_repository "github.com/tribalwarshelp/dcbot/server/repository"

	"github.com/tribalwarshelp/shared/mode"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pgext"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	commandPrefix = "tw!"
)

var status = "Tribal Wars | " + discord.HelpCommand.WithPrefix(commandPrefix).String()

func init() {
	os.Setenv("TZ", "UTC")

	if mode.Get() == mode.DevelopmentMode {
		godotenv.Load(".env.development")
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err := message.LoadMessageFiles(path.Join(dir, "message", "translations")); err != nil {
		log.Fatal(err)
	}

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
	if strings.ToUpper(os.Getenv("LOG_DB_QUERIES")) == "TRUE" {
		db.AddQueryHook(pgext.DebugHook{
			Verbose: true,
		})
	}

	serverRepo, err := server_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	groupRepo, err := group_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	observationRepo, err := observation_repository.NewPgRepo(db)
	if err != nil {
		log.Fatal(err)
	}

	api := sdk.New(os.Getenv("API_URL"))

	sess, err := discord.New(discord.SessionConfig{
		Token:                 os.Getenv("BOT_TOKEN"),
		CommandPrefix:         commandPrefix,
		Status:                status,
		ObservationRepository: observationRepo,
		ServerRepository:      serverRepo,
		GroupRepository:       groupRepo,
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
		GroupRepo:       groupRepo,
		API:             api,
		Status:          status,
	})
	c.Start()
	defer c.Stop()

	log.Print("Bot is waiting for your actions!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	log.Print("shutting down")
}
