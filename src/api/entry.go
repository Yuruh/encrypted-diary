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

type Url struct {
	ovh.ObjectTempPublicUrl
	entryIdx int
	labelIdx int
	error
}

// For user queries: can include vs must include
func GetEntries(c echo.Context) error {
	var user database.User = c.Get("user").(database.User)

	limit, page, offset, err := paginate.GetPaginationParams(10, c)

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

	if err != nil {
		return InternalError(c, err)
	}

	type Data struct {
		labels []database.Label
		idx int
	}
	var results = 0
	ch := make(chan Data)

	for entryIdx, entry := range entries {
		results++
		go func(idx int, labels []database.Label) {
			ch <- Data{
				labels: PopulateLabelsUrls(labels),
				idx: idx,
			}

		}(entryIdx, entry.Labels)
	}
	for i := 0; i < results; i++ {
		data := <-ch
		entries[data.idx].Labels = data.labels
	}
	
	pagination, err := paginate.GetPaginationResults("entries", uint(limit), uint(page), database.GetDB().Where("user_id = ?", user.ID))
	if err != nil {
		return InternalError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"entries": entries, "pagination": pagination})
}

/*
	Returns a specific entry and the next / prev ones.

	We may want that the prev and next fit search criteria,
	in which case they should be sent here with the same format as in GetEntries
 */
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

	entry.Labels = PopulateLabelsUrls(entry.Labels)

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
		return InternalError(context, err)
	}

	return context.JSON(http.StatusCreated, map[string]interface{}{"entry": entry})
}

/*
	Reads body, and add requested labels in entry
 */
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
		return InternalError(context, err)
	}

	return context.JSON(http.StatusOK, map[string]interface{}{"entry": builtEntry})
}

func DeleteEntry(context echo.Context) error {
	emptyEntry := database.Entry{}
	return DeleteAbstract(context, &emptyEntry)
}

