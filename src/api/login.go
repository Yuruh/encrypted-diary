package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
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

// Add a user to claims without otp secret and password.
// Expects a duration in nanoseconds
// Could implement a key rotation system
func BuildJwtToken(user database.User, sessionDuration time.Duration, secret []byte) string {
	user.Password = ""
	user.OTPSecret = ""

	claims := &TokenClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + int64(sessionDuration / time.Second),
			IssuedAt: time.Now().Unix(),
			Issuer: "auth.yuruh.fr", // This would make sense if auth server was external
			Audience: "api.diary.yuruh.fr",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, _ := token.SignedString(secret)

	return ss
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
		//2FA enabled
		if user.HasRegisteredOTP {
			var active database.TwoFactorsCookie
			cookie, err := context.Cookie("tfa-active")
			// Cookie found
			if err == nil {
				result := database.GetDB().
					Where("user_id = ?", user.ID).
					Where("uuid = ?", cookie.Value).
					Find(&active)
				if !result.RecordNotFound() {
					if time.Now().UnixNano() > active.Expires.UnixNano() {
						err = active.Delete()
						if err != nil {
							sentry.CaptureException(err)
						}
					} else {
						active.LastUsed = time.Now()
						err = database.Update(&active)
						if err != nil {
							sentry.CaptureException(err)
						}
						ss := BuildJwtToken(user, parsedBody.SessionDurationMs, []byte(os.Getenv("ACCESS_TOKEN_SECRET")))
						return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "two_factors_methods": nil})
					}
				}
			}
			methods := [1]string{"OTP"}
			// We generate a token that cannot be used to authenticate request but will be used to validate 2FA
			ss := BuildJwtToken(user, parsedBody.SessionDurationMs, []byte(os.Getenv("2FA_TOKEN_SECRET")))
			return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "two_factors_methods": methods})
		}
		ss := BuildJwtToken(user, parsedBody.SessionDurationMs, []byte(os.Getenv("ACCESS_TOKEN_SECRET")))
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
			return InternalError(context, err)
		}
	}

	if err != nil {
		//		if errors.Is(err, database.ValidationError{}) {
		// not sure how to check which error it is from golang
		return context.String(http.StatusBadRequest, err.Error())
	}
	return context.JSON(http.StatusCreated, map[string]interface{}{"user": user})
}

