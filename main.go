package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/tribalwarshelp/dcbot/message"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	_cron "github.com/tribalwarshelp/dcbot/cron"
	"github.com/tribalwarshelp/dcbot/discord"
	group_repository "github.com/tribalwarshelp/dcbot/group/repository"
	observation_repository "github.com/tribalwarshelp/dcbot/observation/repository"
	server_repository "github.com/tribalwarshelp/dcbot/server/repository"

	"github.com/tribalwarshelp/shared/mode"

	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	commandPrefix = "tw!"
)

var status = "tribalwarshelp.com | " + discord.HelpCommand.WithPrefix(commandPrefix).String()

func init() {
	os.Setenv("TZ", "UTC")

	if mode.Get() == mode.DevelopmentMode {
		godotenv.Load(".env.development")
		logrus.SetLevel(logrus.DebugLevel)
	}

	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	dirWithMessages := path.Join(dir, "message", "translations")
	if err := message.LoadMessageFiles(dirWithMessages); err != nil {
		logrus.Fatal(err)
	}
	logrus.WithField("dir", dirWithMessages).
		WithField("languages", message.LanguageTags()).
		Info("Loaded messages")

	dbOptions := &pg.Options{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		Addr:     os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
	}
	dbFields := logrus.Fields{
		"user":     dbOptions.User,
		"database": dbOptions.Database,
		"addr":     dbOptions.Addr,
	}
	db := pg.Connect(dbOptions)
	defer func() {
		if err := db.Close(); err != nil {
			logrus.WithFields(dbFields).Fatalln(err)
		}
	}()
	if strings.ToUpper(os.Getenv("LOG_DB_QUERIES")) == "TRUE" {
		db.AddQueryHook(pgdebug.DebugHook{
			Verbose: true,
		})
	}

	serverRepo, err := server_repository.NewPgRepo(db)
	if err != nil {
		logrus.Fatal(err)
	}
	groupRepo, err := group_repository.NewPgRepo(db)
	if err != nil {
		logrus.Fatal(err)
	}
	observationRepo, err := observation_repository.NewPgRepo(db)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.WithFields(dbFields).Info("Connected to the database")

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
		logrus.Fatal(err)
	}
	defer sess.Close()
	logrus.WithFields(logrus.Fields{
		"api":           os.Getenv("API_URL"),
		"commandPrefix": commandPrefix,
		"status":        status,
	}).Info("Initialized new Discord session")

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
	logrus.Info("Started the cron scheduler")

	logrus.Info("Bot is running!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	logrus.Info("shutting down")
}
