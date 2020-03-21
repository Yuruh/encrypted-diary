package src

import (
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/Yuruh/encrypted-diary/src/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Handler
func GetEntries(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"entries": []int{}})
}

func AddEntry(context echo.Context) error {
	body := helpers.ReadBody(context.Request().Body)

	var data models.Entry

	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		println(err.Error())
		return context.NoContent(http.StatusBadRequest)
	}

	err = data.Validate()
	if err != nil {
		println(err.Error())
		// Should say which field and why
		return context.NoContent(http.StatusBadRequest)
	}

	err = data.Create()
	if err != nil {
		println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}
}