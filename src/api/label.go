package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func GetPaginationParams(defaultLimit int, context echo.Context) (limit int, offset int, err error) {
	limit = defaultLimit
	offset = 0

	if context.QueryParam("limit") != "" {
		limit, err = strconv.Atoi(context.QueryParam("limit"))
		if err != nil {
			return 0, 0, err
		}
	}

	if context.QueryParam("page") != "" {
		page, err := strconv.Atoi(context.QueryParam("page"))
		if err != nil {
			return 0, 0, err
		}
		if page >= 1 {
			offset = limit * (page - 1)
		}
	}
	return limit, offset, nil
}

func GetLabels(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	limit, offset, err := GetPaginationParams(5, context)
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad query parameters")
	}

	var labels []database.Label
	database.GetDB().
		Where("user_id = ?", user.ID).
		Limit(limit).
		Offset(offset).
		Find(&labels)
	return context.JSON(http.StatusOK, map[string]interface{}{"labels": labels})
}

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
