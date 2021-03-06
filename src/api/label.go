package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/api/paginate"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/Yuruh/encrypted-diary/src/object-storage/ovh"
	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

func PopulateLabelsUrls(labels []database.Label) []database.Label {
	chUrl := make(chan Url)
	var results = 0
	for labelIdx, label := range labels {
		if label.HasAvatar == true {
			results++
			go func(lIdx int, label database.Label) {
				url, err := ovh.GetFileTemporaryAccess(getLabelAvatarFileDescriptor(label), TokenToRemainingDuration())
				chUrl <- Url{
					ObjectTempPublicUrl: url,
					entryIdx:            -1,
					labelIdx:            lIdx,
					error: err,
				}
			}(labelIdx, label)
		}
	}
	for i := 0; i < results; i++ {
		url := <-chUrl
		if url.error != nil {
			sentry.CaptureException(url.error)
		}
		labels[url.labelIdx].AvatarUrl = url.URL
	}
	return labels
}

func GetLabels(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	limit, _, offset, err := paginate.GetPaginationParams(5, context)
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad query parameters")
	}

	// Should it be sanitized ?
	name := context.QueryParam("name")

	// We read excluded as a json array. Not sure if it is good practice but theres no standard in url and it seems the easiest way
	excludedMarshall := context.QueryParam("excluded_ids")

	// We initialize with -1 so query uses an impossible value in the NOT IN clause
	var excluded = make([]int, 0)
	if excludedMarshall != "" {
		err = json.Unmarshal([]byte(excludedMarshall), &excluded)
		if err != nil {
			return context.String(http.StatusBadRequest, "Bad query parameters")
		}
	}
	if len(excluded) == 0 {
		excluded = []int{-1}
	}

	var labels []database.Label
	database.GetDB().
		Where("user_id = ?", user.ID).
		Not("id IN (?)", excluded).
		Limit(limit).
		Offset(offset).
		// We use levenshtein https://www.postgresql.org/docs/9.1/fuzzystrmatch.html
		// Note: It seems to be case influenced, so we work on lowercase
		Order(gorm.Expr("levenshtein(LOWER(?), SUBSTRING(LOWER(labels.name), 1, LENGTH(?))) ASC", name, name)).
		Find(&labels)


	labels = PopulateLabelsUrls(labels)

	return context.JSON(http.StatusOK, map[string]interface{}{"labels": labels})
}

// almost the exact same code as add entry, could be refactored but not sure how without generic
// maybe with reflect, https://stackoverflow.com/questions/51097211/how-to-pass-type-to-function-argument
// I feel reflect is a terrible idea, maybe this can be interfaced

// May not be as similar in the end, still the method is too big imo
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

	var existingLabel database.Label
	result := database.GetDB().
		Where("user_id = ?", user.ID).
		Where("LOWER(name) = ?", strings.ToLower(label.Name)).
		Find(&existingLabel)
	if !result.RecordNotFound() {
		return context.String(http.StatusConflict, "Label with name " + label.Name + " already exists")
	}

	err = database.Insert(&label)

	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		return InternalError(context, err)
	}

	return context.JSON(http.StatusCreated, map[string]interface{}{"label": label})
}

func getLabelAvatarFileDescriptor(label database.Label) string {
	return "label_" + strconv.Itoa(int(label.ID)) + "_avatar"
}

func EditLabel(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}
	var label database.Label
	result := database.GetDB().
		Where("ID = ?", id).
		Where("user_id = ?", user.ID).
		First(&label)
	if result.RecordNotFound() {
		return context.String(http.StatusNotFound, "Label not found")
	}

	form, _ := context.FormParams()
	fmt.Println(form)

	// avatar is not in forms, apparently because its a file
	avatar, err := context.FormFile("avatar")
	if err == nil {
		file, err := avatar.Open()
		if err != nil {
			sentry.CaptureException(err)
			return context.String(http.StatusBadRequest, "Could not read avatar")
		}
		err = ovh.UploadFileToPrivateObjectStorage(getLabelAvatarFileDescriptor(label), file)
		if err != nil {
			return InternalError(context, err)
		}
		url, err := ovh.GetFileTemporaryAccess(getLabelAvatarFileDescriptor(label), TokenToRemainingDuration())
		if err != nil {
			return InternalError(context, err)
		}
		label.HasAvatar = true
		label.AvatarUrl = url.URL
	}

	if form.Get("json") != "" {
		body := context.FormValue("json")

		var partialLabel database.PartialLabel

		err = json.Unmarshal([]byte(body), &partialLabel)
		if err != nil {
			return context.String(http.StatusBadRequest, "Could not read JSON body")
		}
		label.PartialLabel = partialLabel
	}

	err = database.Update(&label)

	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}
	if err != nil {
		return InternalError(context, err)

	}

	return context.JSON(http.StatusOK, map[string]interface{}{"label": label})
}

func DeleteLabel(context echo.Context) error {
	label := database.Label{}
	return DeleteAbstract(context, &label)
}
