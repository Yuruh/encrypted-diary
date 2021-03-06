package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgryski/dgoogauth"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestValidateOTPCode(t *testing.T) {
	assert := asserthelper.New(t)
	user, _ := SetupUsers()

	user.OTPSecret = "2SH3V3GDW7ZNMGYE"
	database.Update(&user)


	user.Password = ""
	user.OTPSecret = ""
	claims := &TokenClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 60 * 30,
			IssuedAt: time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("2FA_TOKEN_SECRET")))

	// Bad passcode format test
	body := OTPCodeBody{
		Passcode: "bad passcode",
		Token:    ss,
		KeepActive: false,
	}
	marsh, _ := json.Marshal(body)
	context, recorder := BuildEchoContext(marsh, echo.MIMEApplicationJSON)
	err := ValidateOTPCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// Bad token test
	body.Passcode = "123456"
	body.Token = "bad token"
	marsh, _ = json.Marshal(body)
	context, recorder = BuildEchoContext(marsh, echo.MIMEApplicationJSON)
	err = ValidateOTPCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Equal("Bad Token", recorder.Body.String())

	// Bad code value test
	body.Token = ss
	marsh, _ = json.Marshal(body)
	context, recorder = BuildEchoContext(marsh, echo.MIMEApplicationJSON)
	err = ValidateOTPCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Equal("Code refused", recorder.Body.String())

	// Not sure how to test valid, as the code is generated by app

	body.KeepActive = true;
	// Test valid code
	otpconf := &dgoogauth.OTPConfig{
		Secret:      "2SH3V3GDW7ZNMGYE",
		WindowSize:  2,
		HotpCounter: 0,
	}
	var t0 int64
	if otpconf.UTC {
		t0 = int64(time.Now().UTC().Unix() / 30)
	} else {
		t0 = int64(time.Now().Unix() / 30)
	}
	c := dgoogauth.ComputeCode(otpconf.Secret, t0)
	code := fmt.Sprintf("%06d", c)
	body.Passcode = code
	marsh, _ = json.Marshal(body)
	context, recorder = BuildEchoContext(marsh, echo.MIMEApplicationJSON)
	context.Request().Header.Set("X-FORWARDED-FOR", "69.96.69.96")
	err = ValidateOTPCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	assert.Contains(recorder.Header().Get("set-cookie"), "Expires=")
	assert.Contains(recorder.Header().Get("set-cookie"), "Secure")
	assert.Contains(recorder.Header().Get("set-cookie"), "HttpOnly")
	assert.Contains(recorder.Header().Get("set-cookie"), "tfa-active=")

	var cookie database.TwoFactorsCookie
	database.GetDB().Where("user_id = ?", user.ID).First(&cookie)

	// Greater than 13 days, lower than 15 days
	assert.Greater(cookie.Expires.UnixNano(), time.Now().Add(time.Hour * 24 * 13).UnixNano())
	assert.Greater(time.Now().Add(time.Hour * 24 * 15).UnixNano(), cookie.Expires.UnixNano())
}

func TestRequestGoogleAuthenticatorQRCode(t *testing.T) {
	assert := asserthelper.New(t)
	user, _ := SetupUsers()
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	err := RequestGoogleAuthenticatorQRCode(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("image/png", recorder.Header().Get("content-type"))
	assert.Greater(len(recorder.Body.Bytes()), 600)
	assert.Greater(900, len(recorder.Body.Bytes()))

	var updatedUser database.User
	database.GetDB().Find(&updatedUser)
	assert.Greater(5, len(user.OTPSecret))
}

func TestRequestTwoFactorsToken(t *testing.T) {
	assert := asserthelper.New(t)
	SetupUsers()
	context, recorder := BuildEchoContext(nil, echo.MIMEApplicationJSON)

	err := RequestTwoFactorsToken(context)
	assert.Nil(err)
	assert.Equal(http.StatusOK, recorder.Code)

	var response loginResponse

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	if len(response.Token) < 300 {
		t.Errorf("Token length incorrect, expected at least %v, got %v", 300, len(response.Token))
	}

	if parsedToken, _ := jwt.Parse(response.Token, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("2FA_TOKEN_SECRET"), nil
	}); parsedToken != nil {
		var parsedUser = parsedToken.Claims.(jwt.MapClaims)["user"].(map[string]interface{})
		assert.Equal(parsedUser["email"], UserHasAccessEmail)
		assert.Equal(parsedUser["Password"], nil)
	} else {
		t.Errorf("Could not decde JWT token")
	}

}