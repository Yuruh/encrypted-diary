package api

import (
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type getMeResponse struct {
	User database.User `json:"user"`
}

func TestGetMe(t *testing.T) {
	SetupUsers()
	assert := asserthelper.New(t)

	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	err := GetMe(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response getMeResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal(UserHasAccessEmail, response.User.Email)
}