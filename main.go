package main

import (
	"errors"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/api"
	"github.com/Yuruh/encrypted-diary/src/database"
	"os"
)

func ensureEnvSet() error {
	required := []string{
		"ACCESS_TOKEN_SECRET",
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
	err := ensureEnvSet()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer database.GetDB().Close()

	database.RunMigration()

	api.RunHttpServer()
}


func Dummy() int8 {
	return 1
}
