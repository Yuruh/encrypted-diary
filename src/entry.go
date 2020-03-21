package src

import (
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Handler
func GetEntries(c echo.Context) error {
	var entries []database.Entry
	database.GetDB().Find(&entries)
	return c.JSON(http.StatusOK, map[string]interface{}{"entries": entries})
}

func AddEntry(context echo.Context) error {
	body := helpers.ReadBody(context.Request().Body)

	var entry database.Entry

	err := json.Unmarshal([]byte(body), &entry)
	if err != nil {
		println(err.Error())
		return context.String(http.StatusBadRequest, "Could not read JSON body")
	}

	err = database.Insert(entry)

	if err != nil {
//		if errors.Is(err, database.ValidationError{}) {
		// not sure how to check which error it is from golang
		return context.String(http.StatusBadRequest, err.Error())
//		}
		println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}
	return context.NoContent(http.StatusCreated)
}