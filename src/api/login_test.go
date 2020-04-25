package api

import (
	"bytes"
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	t.Run("User found", caseRequireMoreThan1hours)
	t.Run("User found", caseRequire30Min)


	db := database.GetDB().Delete(&user)
	if db.Error != nil {
		t.Fatal(db.Error.Error())
	}
}

func caseUserNotFound(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "doesnt@exists.fr",
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
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Bad status, expected %v, got %v", http.StatusNotFound, recorder.Code)
	}
}

func caseRequireMoreThan1hours(t* testing.T) {
	assert := asserthelper.New(t)

	user := LoginBody{
		Email:     "does@exists.fr",
		Password:  "password",
		SessionDurationMs: 7200000, // 2 hours
	}
	marshalled, _ := json.Marshal(user)
	context, recorder := BuildEchoContext(marshalled, echo.MIMEApplicationJSON)

	err := Login(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response loginResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	if len(response.Token) < 300 {
		t.Errorf("Token length incorrect, expected at least %v, got %v", 300, len(response.Token))
	}
	// decode JWT token without verifying the signature
	if token, _ := jwt.Parse(response.Token, nil); token != nil {
		// Assert that it expires after now
		assert.Equal(true, token.Claims.(jwt.MapClaims).VerifyExpiresAt(time.Now().Unix(), true))
		// Assert that it expires before 1h
		assert.Equal(false, token.Claims.(jwt.MapClaims).VerifyExpiresAt(int64(float64(time.Now().Unix()) + (float64(time.Hour) * 1.1 / float64(time.Second))), true))
	} else {
		t.Errorf("Could not decde JWT token")
	}
}

func caseRequire30Min(t* testing.T) {
	assert := asserthelper.New(t)

	user := LoginBody{
		Email:     "does@exists.fr",
		Password:  "password",
		SessionDurationMs: 7200000 / 4, // 30 minutes
	}
	marshalled, _ := json.Marshal(user)
	context, recorder := BuildEchoContext(marshalled, echo.MIMEApplicationJSON)

	err := Login(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response loginResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	if len(response.Token) < 300 {
		t.Errorf("Token length incorrect, expected at least %v, got %v", 300, len(response.Token))
	}
	// decode JWT token without verifying the signature
	if token, _ := jwt.Parse(response.Token, nil); token != nil {
		// Assert that it expires after now
		assert.Equal(true, token.Claims.(jwt.MapClaims).VerifyExpiresAt(time.Now().Unix(), true))
		// Assert that it expires before a bit more than 30 min
		assert.Equal(true, token.Claims.(jwt.MapClaims).VerifyExpiresAt(int64(float64(time.Now().Unix()) + (float64(time.Hour) * 0.2 / float64(time.Second))), true))

		// Assert that it expires before a bit more than 30 min
		assert.Equal(false, token.Claims.(jwt.MapClaims).VerifyExpiresAt(int64(float64(time.Now().Unix()) + (float64(time.Hour) * 0.6 / float64(time.Second))), true))
	} else {
		t.Errorf("Could not decde JWT token")
	}
}

type loginResponse struct {
	Token string `json:"token"`
	TwoFactorsMethods []string `json:"two_factors_methods"`
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
	if len(response.Token) < 300 {
		t.Errorf("Token length incorrect, expected at least %v, got %v", 300, len(response.Token))
	}
}

func TestRegister(t *testing.T) {
	database.GetDB().Where("email = ?", "does@exists.fr").Unscoped().Delete(database.User{})
	database.GetDB().Where("email = ?", "doesnt@exists.fr").Unscoped().Delete(database.User{})
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	user := database.User{
		Email:     "does@exists.fr",
		Password:  string(hash),
	}

	err := database.Insert(&user)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Run("User already exists", caseUserAlreadyExists)
	t.Run("Bad email format", caseBadEmailFormat)
	t.Run("Bad pwd format", caseBadPasswordFormat)
	t.Run("User created", caseUserCreated)
	db := database.GetDB().Delete(&user)
	if db.Error != nil {
		t.Fatal(db.Error.Error())
	}
}

func caseUserAlreadyExists(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "does@exists.fr",
		Password:  "Aze@ty1iop",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Register(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusConflict {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusConflict, recorder.Code, recorder.Body.String())
	}
}

func caseBadEmailFormat(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "badmail",
		Password:  "Aze@ty1iop",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Register(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func caseBadPasswordFormat(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "test@exists.fr",
		Password:  "azertyuiop",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Register(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

func caseUserCreated(t *testing.T) {
	e := echo.New()
	user := LoginBody{
		Email:     "doesnt@exists.fr",
		Password:  "Aze@ty1iop",
	}
	marshalled, _ := json.Marshal(user)
	request, err := http.NewRequest("POST", "/register", bytes.NewReader(marshalled))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = Register(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestRequestGoogleAuthenticatorQRCode(t *testing.T) {
	assert := asserthelper.New(t)
	user, _ := SetupUsers()
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	context.Set("user", user)

	err := RequestGoogleAuthenticatorQRCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("image/png", recorder.Header().Get("content-type"))
	assert.Greater(len(recorder.Body.Bytes()), 800)
	assert.Greater(1200, len(recorder.Body.Bytes()))

	var updatedUser database.User
	database.GetDB().Find(&updatedUser)
	assert.Greater(5, len(user.OTPSecret))
}


func TestLoginRequestOTP(t *testing.T) {
	assert := asserthelper.New(t)
	user, _ := SetupUsers()

	user.OTPSecret = "a very secret otp"
	database.Update(&user)

	body := LoginBody{
		Email:     UserHasAccessEmail,
		Password:  "azer",
		SessionDurationMs: 7200000 / 4, // 30 minutes
	}
	marsh, _ := json.Marshal(body)
	context, recorder := BuildEchoContext(marsh, echo.MIMEApplicationJSON)

	err := Login(context)
	assert.Nil(err)

	var response loginResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)

	assert.Equal(1, len(response.TwoFactorsMethods))
	assert.Equal("OTP", response.TwoFactorsMethods[0])
	assert.Greater(len(response.Token), 400)
}