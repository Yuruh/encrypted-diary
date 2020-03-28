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

// Handler
func GetEntries(c echo.Context) error {
	var user database.User = c.Get("user").(database.User)

	var limit = 10
	var offset = 0
	var err error

	if c.QueryParam("limit") != "" {
		limit, err = strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Bad query param 'limit', expected number")
		}
	}

	if c.QueryParam("page") != "" {
		page, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Bad query param 'page', expected number")
		}
		if page >= 2 {
			offset = limit * (page - 2)
		}
	}

	var entries []database.Entry
	database.GetDB().
		Where("user_id = ?", user.ID).
		Select("id, title, updated_at, created_at").
		Order("updated_at desc").
		Limit(limit).
		Offset(offset).
		Find(&entries)
	return c.JSON(http.StatusOK, map[string]interface{}{"entries": entries})
}

func GetEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var entry database.Entry
	result := database.GetDB().
		Where("ID = ?", id).
		Where("user_id = ?", user.ID).
		First(&entry)
	if result.RecordNotFound() {
		return context.String(http.StatusNotFound, "Entry not found")
	}
	return context.JSON(http.StatusOK, map[string]interface{}{"entry": entry})
}

func AddEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	body := helpers.ReadBody(context.Request().Body)

	var partialEntry database.PartialEntry

	err := json.Unmarshal([]byte(body), &partialEntry)
	if err != nil {
		println(err.Error())
		return context.String(http.StatusBadRequest, "Could not read JSON body")
	}

	var entry = database.Entry{
		PartialEntry: partialEntry,
		UserID:user.ID,
	}

	err = database.Insert(&entry)


	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		fmt.Println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}

	return context.JSON(http.StatusCreated, map[string]interface{}{"entry": entry})
}

func EditEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var entry database.Entry
	result := database.GetDB().
		Where("ID = ?", id).
		Where("user_id = ?", user.ID).
		First(&entry)
	if result.RecordNotFound() {
		return context.String(http.StatusNotFound, "Entry not found")
	}

	body := helpers.ReadBody(context.Request().Body)

	var partialEntry database.PartialEntry

	err = json.Unmarshal([]byte(body), &partialEntry)
	if err != nil {
		return context.String(http.StatusBadRequest, "Could not read JSON body")
	}
	entry.PartialEntry = partialEntry

	err = database.Update(&entry)
	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		fmt.Println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}

	return context.JSON(http.StatusOK, map[string]interface{}{"entry": entry})
}

func DeleteEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var entry database.Entry
	result := database.GetDB().
		Where("ID = ?", id).
		Where("user_id = ?", user.ID).
		First(&entry)
	if result.RecordNotFound() {
		return context.String(http.StatusNotFound, "Entry not found")
	}
	result = database.GetDB().Delete(&entry)
	if result.Error != nil {
		fmt.Println(result.Error.Error())
		return context.NoContent(http.StatusInternalServerError)
	}
	return context.NoContent(http.StatusOK)
}