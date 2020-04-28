package main

import (
	"errors"
	"github.com/Yuruh/encrypted-diary/src/api"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/getsentry/sentry-go"
	"log"
	"math/rand"
	"os"
	"time"
)

func ensureEnvSet() error {
	required := []string{
		"ACCESS_TOKEN_SECRET",
		"2FA_TOKEN_SECRET",
		"DIARY_DB_USER",
		"DIARY_DB_PWD",
	}
	for _, elem := range required {
		if os.Getenv(elem) == "" {
			return errors.New("Env variable " + elem + " missing")
		}
	}

	return nil
}

func main() {
	rand.Seed(int64(os.Getpid()) * time.Now().Unix())
	err := ensureEnvSet()
	if err != nil {
		sentry.CaptureException(err)
		os.Exit(1)
	}
	err = sentry.Init(sentry.ClientOptions{})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

/*	eventId := sentry.CaptureMessage("It works!")
	log.Println("Sentry event id", *eventId)
	eventId = sentry.CaptureException(errors.New("sentry test error"))
	log.Println("Sentry error event id", *eventId)*/

	defer database.GetDB().Close()

	database.RunMigration()

	api.RunHttpServer()
}


func Dummy() int8 {
	return 1
}
