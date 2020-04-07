package api

import (
	"bytes"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
)

const (
	UserHasAccessEmail = "user1@user.com"
	UserNoAccessEmail = "user2@user.com"
)

func SetupUsers() (database.User, database.User) {
	err := database.GetDB().Unscoped().Delete(database.User{})
	if err.Error != nil {
		fmt.Println(err.Error.Error())
	}
	var user1 = database.User{
		BaseModel: database.BaseModel{},
		Email:     UserHasAccessEmail,
		Password:  "azer",
	}
	_ = database.Insert(&user1)
	var user2 = database.User{
		BaseModel: database.BaseModel{},
		Email:     UserNoAccessEmail,
		Password:  "azer",
	}
	_ = database.Insert(&user2)
	return user1, user2
}

func BuildEchoContext(body []byte) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	request, _ := http.NewRequest("POST", "/entries", bytes.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	return context, recorder
}