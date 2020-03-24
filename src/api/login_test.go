package api

import (
	"bytes"
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	database.GetDB().Where("email = ?", "does@exists.fr").Unscoped().Delete(database.User{})
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	user := database.User{
		Email:     "does@exists.fr",
		Password:  string(hash),
	}

	err := database.Insert(&user)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Run("User not found", caseUserNotFound)
	t.Run("User found", caseUserFound)
	db := database.GetDB().Delete(&user)
	if db.Error != nil {
		t.Fatal(db.Error.Error())
	}
}

func caseUserNotFound(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "doesnt@exists.fr",
		Password:  "azertyuiop",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/login", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Login(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Bad status, expected %v, got %v", http.StatusNotFound, recorder.Code)
	}
}

type loginResponse struct {
	Token string `json:"token"`
}

func caseUserFound(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "does@exists.fr",
		Password:  "password",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/login", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Login(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}
	var response loginResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}
	if len(response.Token) != 305 {
		t.Errorf("Token length incorrect, expected %v, got %v", 305, len(response.Token))
	}
}

func TestRegister(t *testing.T) {
	t.Run("User already exists", caseUserAlreadyExists)
	t.Run("Bad parameters", caseRegisterBadParameters)
	t.Run("User created", caseUserCreated)
}

func caseUserAlreadyExists(t *testing.T) {

}

func caseRegisterBadParameters(t *testing.T) {

}

func caseUserCreated(t *testing.T) {

}