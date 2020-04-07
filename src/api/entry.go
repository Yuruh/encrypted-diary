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
	var ret map[string]interface{} = map[string]interface{}{"entry": entry}
	var nextEntry database.Entry
	result = database.GetDB().
		Where("user_id = ?", user.ID).
		Order("created_at asc").
		Where("created_at > ?", entry.CreatedAt).
		First(&nextEntry)
	if !result.RecordNotFound() {
		ret["next_entry_id"] = nextEntry.ID
	}
	var prevEntry database.Entry
	result = database.GetDB().
		Where("user_id = ?", user.ID).
		Order("created_at desc").
		Where("created_at < ?", entry.CreatedAt).
		First(&prevEntry)
	if !result.RecordNotFound() {
		ret["prev_entry_id"] = prevEntry.ID
	}

	// todo refacto, or as an exercise, build this as a single data base request;

	return context.JSON(http.StatusOK, ret)
}

type AddEntryRequestBody struct {
	database.PartialEntry
	LabelsID []uint `json:"labels_id"`
}

func AddEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	entry, errorString := buildEntryFromRequestBody(context, user)
	if errorString != "" {
		return context.String(http.StatusBadRequest, errorString)
	}
	err := database.Insert(&entry)

	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		fmt.Println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}

	return context.JSON(http.StatusCreated, map[string]interface{}{"entry": entry})
}

func buildEntryFromRequestBody(context echo.Context, user database.User) (database.Entry, string) {
	body := helpers.ReadBody(context.Request().Body)

	var requestBody AddEntryRequestBody

	err := json.Unmarshal([]byte(body), &requestBody)
	if err != nil {
		return database.Entry{}, "Could not read JSON body"
	}

	// request to find all users label in labels_id
	var labels []database.Label

	response := database.GetDB().
		Where("user_id = ?", user.ID).
		Where("id IN (?)", requestBody.LabelsID).
		Find(&labels)
	if response.Error != nil {
		fmt.Println(response.Error.Error())
	}

	return database.Entry{
		PartialEntry: requestBody.PartialEntry,
		UserID:user.ID,
		Labels: labels,
	}, ""
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

	builtEntry, errorString := buildEntryFromRequestBody(context, user)
	if errorString != "" {
		return context.String(http.StatusBadRequest, errorString)
	}
	builtEntry.ID = entry.ID

	err = database.Update(&builtEntry)
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