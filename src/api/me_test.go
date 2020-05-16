package api

import (
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type getMeResponse struct {
	User database.User `json:"user"`
}

func TestGetMe(t *testing.T) {
	user, _ := SetupUsers()
	assert := asserthelper.New(t)

	generatedUuid := uuid.New().String()
	validCookie := database.TwoFactorsCookie{
		Uuid:      generatedUuid,
		IpAddr:    "1.2.4.4",
		UserAgent: "bond",
		Expires:   time.Now().Add(time.Hour * 4),
		UserID:	   user.ID,
		LastUsed:  time.Now(),
	}
	validCookie.Create()

	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	err := GetMe(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response getMeResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal(UserHasAccessEmail, response.User.Email)
	if assert.Equal(1, len(response.User.TwoFactorsCookies)) {
		assert.Equal(generatedUuid, response.User.TwoFactorsCookies[0].Uuid)
	}
}