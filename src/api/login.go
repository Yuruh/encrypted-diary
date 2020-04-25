package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/authentication"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"unicode"
)

// We use login body instead of user because user json blocks password
type LoginBody struct {
	Email string `json:"email"`
	Password string `json:"password" validate:"min=9"`
	SessionDurationMs time.Duration `json:"session_duration_ms"`
	TwoFactorsCookie string `json:"two_factors_cookie"`
}
/*
"^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$"

Can be used client side to validate password
 */

const maxTokenDuration = time.Hour * 1
const defaultTokenDuration = time.Minute * 30
/*
	COOKIE 2FA

	CLIENT LOG in without providing 2FA cookie or invalid / expired cookie --->
						<-- Server says give me 2FA
	CLIENT sends: UUID, OTP, -->
						<-- Server stores it, sets an expiration date (even if client sets one), stores user-agent, stores ip

 */

func sanitizeParsedBody(parsedBody *LoginBody) {
	if parsedBody.SessionDurationMs != 0 {
		// so we can use GO's time.Duration, as nanoseconds
		parsedBody.SessionDurationMs = parsedBody.SessionDurationMs * time.Millisecond
	} else {
		parsedBody.SessionDurationMs = defaultTokenDuration
	}

	if parsedBody.SessionDurationMs > maxTokenDuration {
		parsedBody.SessionDurationMs = maxTokenDuration
	}
}

// 2FA could check: Cookie, IP, device / fingerprint, link to element which validates the cookie
func Login(context echo.Context) error {
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody LoginBody

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return context.NoContent(http.StatusBadRequest)
	}
	sanitizeParsedBody(&parsedBody)

	var user database.User
	database.GetDB().Where("email = ?", parsedBody.Email).First(&user)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(parsedBody.Password))

	if err != nil {
		return context.String(http.StatusNotFound, "User not found")
	} else {
		user.Password = ""
		// todo check that otpsecret does not appear in token
		claims := &TokenClaims{
			user,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Unix() + int64(parsedBody.SessionDurationMs / time.Second),
				IssuedAt: time.Now().Unix(),
				Issuer: "auth.yuruh.fr", // This would make sense if auth server was external
				Audience: "api.diary.yuruh.fr",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

		//2FA enabled
		if user.OTPSecret != "" {
			methods := [1]string{"OTP"}
			ss, _ := token.SignedString([]byte(os.Getenv("2FA_TOKEN_SECRET")))
			return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "two_factors_methods": methods})
		}

		ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

		return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "two_factors_methods": nil})
	}
}

func verifyPassword(s string) bool {
	var number, upper, special bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
		default:
			//return false, false, false
		}
	}
	return len(s) >= 8 && number && upper && special
}

func Register(context echo.Context) error {
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody LoginBody

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return context.NoContent(http.StatusBadRequest)
	}

	if !verifyPassword(parsedBody.Password) {
		return context.String(http.StatusBadRequest, "Bad password. Requirements: Minimum eight characters, at least one uppercase letter, " +
			"one lowercase letter, one number and one special character")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(parsedBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return context.String(http.StatusBadRequest, "Could not process password")
	}

	user := database.User{
		Email:     parsedBody.Email,
		Password:  string(hash),
	}
	err = database.Insert(&user)
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code)

		// Are values defined somewhere ?
		if err.Code == "23505" { //https://www.postgresql.org/docs/current/errcodes-appendix.html
			return context.NoContent(http.StatusConflict)
		} else {
			fmt.Println("Unhandled pq error", err.Code, err.Code.Name())
			return context.NoContent(http.StatusInternalServerError)
		}
	}

	if err != nil {
		//		if errors.Is(err, database.ValidationError{}) {
		// not sure how to check which error it is from golang
		return context.String(http.StatusBadRequest, err.Error())
	}
	return context.JSON(http.StatusCreated, map[string]interface{}{"user": user})
}

func RequestGoogleAuthenticatorQRCode(context echo.Context) error {
	var user database.User = context.Get("user").(database.User)
	secret := authentication.GenerateRandomSecret()

	uri := authentication.BuildGAuthURI(user.Email, secret)
	png, err := authentication.GenerateQRCodeFromURI(uri)
	if err != nil {
		return InternalError(context, err)
	}
	user.OTPSecret = secret
	err = database.Update(&user)
	if err != nil {
		return InternalError(context, err)
	}
	return context.Blob(http.StatusOK, "image/png", png)
}

// Return HTTP 500 and log error to sentry
func InternalError(context echo.Context, err error) error {
	sentry.CaptureException(err)
	fmt.Println("INTERNAL ERROR:", err.Error())
	return context.NoContent(http.StatusInternalServerError)
}

func ValidateGoogleAuthCode(context echo.Context) error {
	type Body struct {
		Token string `json:"token" validate:"required,len=6,numeric"`
		Email string `json:"email" validate:"required,email"`
	}

	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody Body
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {

		return context.String(http.StatusBadRequest, "Bad Body")
	}

	validate := validator.New()

	err = validate.Struct(&parsedBody)
	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}

	var user database.User
	dbCpy := database.GetDB().Where("email = ?", parsedBody.Email).Find(&user)
	if dbCpy.RecordNotFound() {
		return context.NoContent(http.StatusNotFound)
	}
	if dbCpy.Error != nil {
		return InternalError(context, dbCpy.Error)
	}

	// todo handle otpsecret encryption
	valid, err := authentication.Authorize(parsedBody.Token, user.OTPSecret)
	if err != nil {
		return InternalError(context, err)
	}
	if valid {
		return context.NoContent(http.StatusOK)
	} else {
		return context.String(http.StatusBadRequest, "Code refused")
	}}
