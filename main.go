package main

import (
	"github.com/Yuruh/encrypted-diary/src/api"
	"github.com/Yuruh/encrypted-diary/src/database"
)

func main() {
	defer database.GetDB().Close()

	database.RunMigration()

	api.RunHttpServer()
}


func Dummy() int8 {
	return 1
}
