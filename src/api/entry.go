package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/api/paginate"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/Yuruh/encrypted-diary/src/object-storage/ovh"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

/*
	Je veux inférrer ce que l'user demande à partir du champs "search"

	Décision: on return uniquement des entrées. (On pourrait choisir de retourner un calendrier, un label, ce genre de trucs)

	Les 3 axes possible: recherche par date, par label ou par titre

	Par titre: full text search ? probablement overkill car les titres ne sont pas long, il doit y avoir plus opti

	Par label: levenshtein, cf requete du GET /label

	Par date: "janvier 2020", "december", "01/02/2020", "01 02 2020",

	c'est le plus compliqué, si je veux respecter plusieurs langages et format de dates

	coté front, avoir un petit (?) pour expliquer les formats et la recherche en général

	et avoir un load moar sur le dropdown de la recherche

	avoir un "indicateur de performance" pour chaque et order par cet indicateur

	possibilité de cumuler les requetes par exemple  "Mars 2020;Games" --> sort en priorité les entrées qui contiennent le label Games pendant le mois de mars 2020
 */

// For user queries: can include vs must include
func GetEntries(c echo.Context) error {
	var user database.User = c.Get("user").(database.User)

	limit, page, offset, err := paginate.GetPaginationParams(10, c)
	fmt.Println(page)

	if err != nil {
		return c.String(http.StatusBadRequest, "Bad query parameters")
	}

	var entries []database.Entry
	err = database.GetDB().
		Preload("Labels").
		Where("user_id = ?", user.ID).
		Select("id, title, updated_at, created_at, LENGTH(title)").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&entries).
		Error

	/*
		TODO: optimize by:
		1/ goroutine
		2/ only one url per label ID

	 */
	// might retrieve several access for same label
	for entryIdx, entry := range entries {
		for idx, label := range  entry.Labels {
			if label.HasAvatar {
				access, err := ovh.GetFileTemporaryAccess(LabelAvatarFileDescriptor(label), TokenToRemainingDuration())
				if err != nil {
					fmt.Println(err.Error())
				} else {
					entries[entryIdx].Labels[idx].AvatarUrl = access.URL
				}
			}
		}
	}


	if err != nil {
		fmt.Println(err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	
	pagination, err := paginate.GetPaginationResults("entries", uint(limit), uint(page), database.GetDB().Where("user_id = ?", user.ID))
	if err != nil {
		fmt.Println(err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"entries": entries, "pagination": pagination})
}

func GetEntry(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var entry database.Entry
	result := database.GetDB().
		Preload("Labels").
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


	for idx, label := range  entry.Labels {
		if label.HasAvatar {
			access, err := ovh.GetFileTemporaryAccess(LabelAvatarFileDescriptor(label), TokenToRemainingDuration())
			if err != nil {
				fmt.Println(err.Error())
			} else {
				entry.Labels[idx].AvatarUrl = access.URL
			}
		}
	}



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

	/*
		We clear all associations before inserting the correct ones.
		This creates two problems:
			* if the update goes wrong, all labels association are lost
			* not optimized
	 */
	database.GetDB().Model(&entry).Association("Labels").Clear()

	err = database.Update(&builtEntry)
	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		fmt.Println(err.Error())
		return context.NoContent(http.StatusInternalServerError)
	}

	return context.JSON(http.StatusOK, map[string]interface{}{"entry": builtEntry})
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