package api

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func TestDeleteAbstract(t *testing.T) {
	assert := asserthelper.New(t)

	user, _ := SetupUsers()

	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "Entry content",
			Title:   "Title",
		},
		UserID: user.ID,
	}
	err := database.Insert(&entry)


	// assert works on entry
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(entry.ID)))

	emptyEntry := database.Entry{}
	err = DeleteAbstract(context, &emptyEntry)

	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var resultEntry database.Entry
	result := database.GetDB().Where("ID = ?", entry.ID).First(&resultEntry)
	assert.Equal(true, result.RecordNotFound())
}

func TestDeleteAbstractLabel(t *testing.T) {
	assert := asserthelper.New(t)

	user, _ := SetupUsers()

	label := database.Label{
		PartialLabel: database.PartialLabel{Name:"label", Color:"#FFFFFF"},
		UserID:       user.ID,
		HasAvatar:    false,
	}
	err := database.Insert(&label)
	assert.Nil(err)

	// assert works on label also
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(label.ID)))

	emptyLabel := database.Label{}
	err = DeleteAbstract(context, &emptyLabel)

	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var resultLabel database.Label
	result := database.GetDB().Where("ID = ?", label.ID).First(&resultLabel)
	assert.Equal(true, result.RecordNotFound())
}

func TestDeleteAbstractBadParam(t *testing.T) {
	assert := asserthelper.New(t)

	user1, user2 := SetupUsers()

	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "Entry content",
			Title:   "Title",
		},
		UserID: user1.ID,
	}
	err := database.Insert(&entry)
	assert.Nil(err)

	// missing param
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	err = DeleteAbstract(context, nil)
	assert.Nil(err)

	assert.Equal(http.StatusBadRequest, recorder.Code)

	// Bad param
	context, recorder = BuildEchoContext(nil, echo.MIMEApplicationJSON)
	context.SetParamNames("id")
	context.SetParamValues("patate")

	err = DeleteAbstract(context, nil)
	assert.Nil(err)

	assert.Equal(http.StatusBadRequest, recorder.Code)

	// Internal, cause nil as param
	context, recorder = BuildEchoContext(nil, echo.MIMEApplicationJSON)
	context.SetParamNames("id")
	context.SetParamValues("123456")

	err = DeleteAbstract(context, nil)
	assert.Nil(err)

	assert.Equal(http.StatusInternalServerError, recorder.Code)

	// Not found cause bad id
	context, recorder = BuildEchoContext(nil, echo.MIMEApplicationJSON)
	context.SetParamNames("id")
	context.SetParamValues("123456")

	emptyEntry := database.Entry{}
	err = DeleteAbstract(context, &emptyEntry)
	assert.Nil(err)

	assert.Equal(http.StatusNotFound, recorder.Code)


	// Not found cause bad user
	context, recorder = BuildEchoContext(nil, echo.MIMEApplicationJSON)
	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(entry.ID)))
	context.Set("user", user2)

	emptyEntry = database.Entry{}
	err = DeleteAbstract(context, &emptyEntry)
	assert.Nil(err)

	assert.Equal(http.StatusNotFound, recorder.Code)

}