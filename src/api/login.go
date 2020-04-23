package api

import (
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
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
}
/*
"^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$"

Can be used client side to validate password
 */

const maxTokenDuration = time.Hour * 1
const defaultTokenDuration = time.Minute * 30

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
	// so we can use GO's time.Duration, as nanoseconds
	if parsedBody.SessionDurationMs != 0 {
		parsedBody.SessionDurationMs = parsedBody.SessionDurationMs * time.Millisecond
	} else {
		parsedBody.SessionDurationMs = defaultTokenDuration
	}

	if parsedBody.SessionDurationMs > maxTokenDuration {
		parsedBody.SessionDurationMs = maxTokenDuration
	}

	var user database.User
	database.GetDB().Where("email = ?", parsedBody.Email).First(&user)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(parsedBody.Password))

	fmt.Println(time.Now().Unix() + int64(parsedBody.SessionDurationMs / time.Second))
	if err != nil {
		return context.String(http.StatusNotFound, "User not found")
	} else {
		user.Password = ""
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
		ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

		return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "user": user})
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
