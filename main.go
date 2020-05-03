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

func EnsureEnvSet() error {
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

func InitSentry() {
	log.Println("Launching Sentry ...")
	err := sentry.Init(sentry.ClientOptions{})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

func main() {
	InitSentry()

	rand.Seed(int64(os.Getpid()) * time.Now().Unix())
	err := EnsureEnvSet()
	if err != nil {
		sentry.CaptureException(err)
		os.Exit(1)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	defer database.GetDB().Close()

	api.RunHttpServer()
}
