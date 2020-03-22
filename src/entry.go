package src

import (
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Handler
func GetEntries(c echo.Context) error {
	var entries []database.Entry
	database.GetDB().Find(&entries)
	return c.JSON(http.StatusOK, map[string]interface{}{"entries": entries})
}

func EditEntry(context echo.Context) error {
	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var entry database.Entry
	result := database.GetDB().First(&entry).
		Where("ID = ?", id)
	if result.RecordNotFound() {
		return context.String(http.StatusNotFound, "Entry not found")
	}

	body := helpers.ReadBody(context.Request().Body)

	var partialEntry database.PartialEntry

	err = json.Unmarshal([]byte(body), &partialEntry)
	if err != nil {
		println(err.Error())
		return context.String(http.StatusBadRequest, "Could not read JSON body")
	}
	entry.PartialEntry = partialEntry

	err = database.Update(entry)

	if err != nil {
		return context.String(http.StatusBadRequest, err.Error())
	}
	return context.NoContent(http.StatusCreated)
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