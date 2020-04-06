package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

// almost the exact same code as add entry, could be refactored but not sure how without generic
// maybe with reflect, https://stackoverflow.com/questions/51097211/how-to-pass-type-to-function-argument
// I feel reflect is a terrible idea, maybe this can be interfaced
func AddLabel(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	body := helpers.ReadBody(context.Request().Body)

	var partialLabel database.PartialLabel

	err := json.Unmarshal([]byte(body), &partialLabel)
	if err != nil {
		return context.String(http.StatusBadRequest, "Could not read JSON body")
	}

	var label = database.Label{
		PartialLabel: partialLabel,
		UserID:       user.ID,
	}

	err = database.Insert(&label)

	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		fmt.Println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}

	return context.JSON(http.StatusCreated, map[string]interface{}{"label": label})
}
