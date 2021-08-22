package main

import (
	"github.com/Kichiyaki/appmode"
	"github.com/Kichiyaki/goutil/envutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/cron"
	"github.com/tribalwarshelp/dcbot/discord"
	grouprepository "github.com/tribalwarshelp/dcbot/group/repository"
	observationrepository "github.com/tribalwarshelp/dcbot/observation/repository"
	serverepository "github.com/tribalwarshelp/dcbot/server/repository"

	"github.com/Kichiyaki/go-pg-logrus-query-logger/v10"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
)

const (
	commandPrefix = "tw!"
	status        = "tribalwarshelp.com | tw!help"
)

func init() {
	os.Setenv("TZ", "UTC")

	if appmode.Equals(appmode.DevelopmentMode) {
		godotenv.Load(".env.local")
	}

	prepareLogger()
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	dirWithMessages := path.Join(dir, "message", "translations")
	if err := message.LoadMessages(dirWithMessages); err != nil {
		logrus.WithField("dir", dirWithMessages).Fatal(err)
	}

	db := pg.Connect(&pg.Options{
		User:     envutil.GetenvString("DB_USER"),
		Password: envutil.GetenvString("DB_PASSWORD"),
		Database: envutil.GetenvString("DB_NAME"),
		Addr:     envutil.GetenvString("DB_HOST") + ":" + os.Getenv("DB_PORT"),
	})
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Fatalln(err)
		}
	}()
	if envutil.GetenvBool("LOG_DB_QUERIES") {
		db.AddQueryHook(querylogger.Logger{
			Log:            logrus.NewEntry(logrus.StandardLogger()),
			MaxQueryLength: 5000,
		})
	}

	serverRepo, err := serverepository.NewPgRepository(db)
	if err != nil {
		logrus.Fatal(err)
	}
	groupRepo, err := grouprepository.NewPgRepo(db)
	if err != nil {
		logrus.Fatal(err)
	}
	observationRepo, err := observationrepository.NewPgRepository(db)
	if err != nil {
		logrus.Fatal(err)
	}

	api := sdk.New(envutil.GetenvString("API_URL"))

	sess, err := discord.New(discord.SessionConfig{
		Token:                 envutil.GetenvString("BOT_TOKEN"),
		CommandPrefix:         commandPrefix,
		Status:                status,
		ObservationRepository: observationRepo,
		ServerRepository:      serverRepo,
		GroupRepository:       groupRepo,
		API:                   api,
	})
	if err != nil {
		logrus.
			WithFields(logrus.Fields{
				"api":           envutil.GetenvString("API_URL"),
				"commandPrefix": commandPrefix,
				"status":        status,
			}).
			Fatal(err)
	}
	defer sess.Close()

	c := cron.New(cron.Config{
		ServerRepo:      serverRepo,
		ObservationRepo: observationRepo,
		Discord:         sess,
		GroupRepo:       groupRepo,
		API:             api,
		Status:          status,
	})
	c.Start()
	defer c.Stop()

	logrus.Info("The bot is up and running!")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	logrus.Info("shutting down...")
}

func prepareLogger() {
	if appmode.Equals(appmode.DevelopmentMode) {
		logrus.SetLevel(logrus.DebugLevel)
	}

	timestampFormat := "2006-01-02 15:04:05"
	if appmode.Equals(appmode.ProductionMode) {
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
