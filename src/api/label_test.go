package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type addLabelResponse struct {
	Label database.Label `json:"label"`
}

func TestAddLabel(t *testing.T) {
	user1, _ := SetupUsers()
	assert := asserthelper.New(t)
	marshall, _ := json.Marshal(database.PartialLabel{
		Name:  "Work",
		Color: "#ff00aa",
	})
	context, recorder := BuildEchoContext(marshall, echo.MIMEApplicationJSON)

	fmt.Println("add label 1")
	err := AddLabel(context)

	assert.Nil(err)
	assert.Equal(http.StatusCreated, recorder.Code)

	var response addLabelResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal("Work", response.Label.Name)
	assert.Equal("#ff00aa", response.Label.Color)
	assert.Equal(user1.ID, response.Label.UserID)
}

func TestAddLabelAlreadyExists(t *testing.T) {
	assert := asserthelper.New(t)

	user1, _ := SetupUsers()
	database.GetDB().Create(&database.Label{
		PartialLabel: database.PartialLabel{
			Name: "work",
			Color: "#FF00AA",
		},
		UserID:       user1.ID,
	})
	marshall, _ := json.Marshal(database.PartialLabel{
		Name:  "Work",
		Color: "#ff00aa",
	})
	context, recorder := BuildEchoContext(marshall, echo.MIMEApplicationJSON)

	err := AddLabel(context)
	assert.Nil(err)
	assert.Equal(http.StatusConflict, recorder.Code)
}

func TestAddLabelBadFormat(t *testing.T) {
	assert := asserthelper.New(t)

	marshall, _ := json.Marshal(database.PartialLabel{
		Name:  "Wor&k",
		Color: "#ff00aa",
	})
	context, recorder := BuildEchoContext(marshall, echo.MIMEApplicationJSON)

	err := AddLabel(context)
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, recorder.Code, recorder.Body.String())
}

type getLabelsResponse struct {
	Labels []database.Label `json:"labels"`
}

func TestGetLabelsWithExcluded(t *testing.T) {
	user1, _ := SetupUsers()
	assert := asserthelper.New(t)

	work := database.Label{
		PartialLabel: database.PartialLabel{
			Name:  "Work",
			Color: "#fff",
		},
		UserID: user1.ID,
	}
	family := database.Label{
		PartialLabel: database.PartialLabel{
			Name:  "Family",
			Color: "#fff",
		},
		UserID: user1.ID,

	}
	love := database.Label{
		PartialLabel: database.PartialLabel{
			Name:  "Love",
			Color: "#fff",
		},
		UserID: user1.ID,
	}

	database.GetDB().Create(&work)
	database.GetDB().Create(&family)
	database.GetDB().Create(&love)


	context, recorder := BuildEchoContext([]byte(""), echo.MIMEApplicationJSON)
	excluded := []uint{family.ID, work.ID}
	marshalled, _ := json.Marshal(excluded)
	context.QueryParams().Set("excluded_ids", string(marshalled))

	err := GetLabels(context)
	assert.Equal(http.StatusOK, recorder.Code)

	var response getLabelsResponse
	assert.Nil(err)

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal(1, len(response.Labels))
	assert.Equal("Love", response.Labels[0].Name)
}

func TestGetLabels(t *testing.T) {
	user1, _ := SetupUsers()
	assert := asserthelper.New(t)

	for i := 0; i < 8; i++ {
		database.GetDB().Create(&database.Label{
			PartialLabel: database.PartialLabel{
				Name: "Label " + strconv.Itoa(i),
				Color: "#FF00AA",
			},
			UserID:       user1.ID,
		})
	}

	database.GetDB().Create(&database.Label{
		PartialLabel: database.PartialLabel{
			Name: "Patate",
			Color: "#FF00AA",
		},
		UserID:       user1.ID,
	})

	context, recorder := BuildEchoContext([]byte(""), echo.MIMEApplicationJSON)
	context.QueryParams().Set("name", "p")

	err := GetLabels(context)
	assert.Equal(http.StatusOK, recorder.Code)

	var response getLabelsResponse
	assert.Nil(err)

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal(5, len(response.Labels))
	assert.Equal("Patate", response.Labels[0].Name)
}

func TestEditLabel(t *testing.T) {
	assert := asserthelper.New(t)

	user1, _ := SetupUsers()
	var label database.Label = database.Label{
		PartialLabel: database.PartialLabel{
			Name: "work",
			Color: "#FF00AA",
		},
		UserID:       user1.ID,
	}
	database.GetDB().Create(&label)
	marshall, _ := json.Marshal(database.PartialLabel{
		Name:  "Family",
		Color: "#ff00aa",
	})

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	var fw io.Writer

	fw, err := w.CreateFormField("json")

	io.Copy(fw, strings.NewReader(string(marshall)))
	w.Close()

	context, recorder := BuildEchoContext(b.Bytes(), w.FormDataContentType())

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(label.ID)))

	err = EditLabel(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response addLabelResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal("Family", response.Label.Name)
	assert.Equal("#ff00aa", response.Label.Color)
	assert.Equal(user1.ID, response.Label.UserID)
}

func TestDeleteLabel(t *testing.T) {
	assert := asserthelper.New(t)

	user1, _ := SetupUsers()
	var label database.Label = database.Label{
		PartialLabel: database.PartialLabel{
			Name: "work",
			Color: "#FF00AA",
		},
		UserID:       user1.ID,
	}
	database.GetDB().Create(&label)

	context, recorder := BuildEchoContext([]byte(""), echo.MIMEApplicationJSON)

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(label.ID)))

	err := DeleteLabel(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)


	res := database.GetDB().Find(&label)
	assert.Equal(true, res.RecordNotFound())
}