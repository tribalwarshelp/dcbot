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

	gopglogrusquerylogger "github.com/Kichiyaki/go-pg-logrus-query-logger/v10"
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
		godotenv.Load(".env.local")
	}

	setupLogger()
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	dirWithMessages := path.Join(dir, "message", "translations")
	if err := message.LoadMessages(dirWithMessages); err != nil {
		logrus.Fatal(err)
	}
	logrus.WithField("dir", dirWithMessages).
		WithField("languages", message.LanguageTags()).
		Info("Loaded messages")

	db := pg.Connect(&pg.Options{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		Addr:     os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
	})
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Fatalln(err)
		}
	}()
	if strings.ToUpper(os.Getenv("LOG_DB_QUERIES")) == "TRUE" {
		db.AddQueryHook(gopglogrusquerylogger.QueryLogger{
			Log:            logrus.NewEntry(logrus.StandardLogger()),
			MaxQueryLength: 5000,
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
	}).Info("The Discord session has been initialized")

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

func setupLogger() {
	if mode.Get() == mode.DevelopmentMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	timestampFormat := "2006-01-02 15:04:05"
	if mode.Get() == mode.ProductionMode {
		customFormatter := new(logrus.JSONFormatter)
		customFormatter.TimestampFormat = timestampFormat
		logrus.SetFormatter(customFormatter)
	} else {
		customFormatter := new(logrus.TextFormatter)
		customFormatter.TimestampFormat = timestampFormat
		customFormatter.FullTimestamp = true
		logrus.SetFormatter(customFormatter)
	}
}
