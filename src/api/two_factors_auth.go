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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func RequestGoogleAuthenticatorQRCode(context echo.Context) error {
	var user = context.Get("user").(database.User)
	secret := authentication.GenerateRandomSecret()

	uri := authentication.BuildGAuthURI(user.Email, secret)
	png, err := authentication.GenerateQRCodeFromURI(uri)
	if err != nil {
		return InternalError(context, err)
	}
	err = database.GetDB().Model(&user).Update("OTPSecret", secret).Error
	if err != nil {
		return InternalError(context, err)
	}

	fmt.Println("otp secret in request", user.OTPSecret)


	/*
			This code blocks works but it seems bad practice to send multipart to a client (uncommon behaviour)


			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			var fw io.Writer


			// Instead of using CreateFormField we set by hand because we need to set the content-type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"`, "token"))
		h.Set("Content-Type", "text/plain")
		fw, _ = w.CreatePart(h)
		_, err = io.Copy(fw, strings.NewReader("the token content"))
		if err != nil {
			return InternalError(context, err)
		}

		fw, _ = w.CreateFormFile("qr", "otp-qr-code.png")
		_, err = io.Copy(fw, bytes.NewReader(png))
		if err != nil {
			return InternalError(context, err)
		}

		w.Close()

		return context.Blob(http.StatusOK, w.FormDataContentType(), b.Bytes())*/

	return context.Blob(http.StatusOK, "image/png", png)
}

func RequestTwoFactorsToken(context echo.Context) error {
	var user = context.Get("user").(database.User)

	user.Password = ""
	user.OTPSecret = ""
	claims := &TokenClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 60 * 30,
			IssuedAt: time.Now().Unix(),
			Issuer: "auth.yuruh.fr",
			Audience: "api.diary.yuruh.fr",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("2FA_TOKEN_SECRET")))
	return context.JSON(http.StatusOK, map[string]interface{}{"token": ss})
}

// Return HTTP 500 and log error to sentry
func InternalError(context echo.Context, err error) error {
	sentry.CaptureException(err)
	fmt.Println("INTERNAL ERROR:", err.Error())
	return context.NoContent(http.StatusInternalServerError)
}

func ValidateJWTToken(token string, secret []byte) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %v", err)
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("could not validate token")
	}
}

type OTPCodeBody struct {
	Passcode string `json:"passcode" validate:"required,len=6,numeric"`
	Token string `json:"token" validate:"required"`
}

func ValidateOTPCode(context echo.Context) error {
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody OTPCodeBody
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad Body")
	}

	validate := validator.New()
	err = validate.Struct(&parsedBody)
	if err, ok := err.(validator.ValidationErrors); ok {
		return context.String(http.StatusBadRequest, database.BuildValidationErrorMsg(err))
	}

	claims, err := ValidateJWTToken(parsedBody.Token, []byte(os.Getenv("2FA_TOKEN_SECRET")))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad Token")
	}

	var user database.User

	var parsedUser  = claims["user"].(map[string]interface{})
	// We fetch the user again as the OTP key shouldn't be in the token (readable by anyone)
	dbCpy := database.GetDB().Where("email = ?", parsedUser["email"]).Find(&user)
	if dbCpy.RecordNotFound() {
		return context.NoContent(http.StatusNotFound)
	}
	if dbCpy.Error != nil {
		return InternalError(context, dbCpy.Error)
	}

	// todo handle otpsecret encryption
	valid, err := authentication.Authorize(parsedBody.Passcode, user.OTPSecret)
	if err != nil {
		return InternalError(context, err)
	}
	if valid {
		if !user.HasRegisteredOTP {
			err = database.GetDB().Model(&user).Update("HasRegisteredOTP", true).Error
			if err != nil {
				return InternalError(context, dbCpy.Error)
			}
		}
		// Retrieve duration in nanoseconds from token
		tokenTTL := time.Duration(claims["exp"].(float64)) * time.Second - time.Duration(time.Now().UnixNano())
		ss := BuildJwtToken(user, tokenTTL, []byte(os.Getenv("ACCESS_TOKEN_SECRET")))
		return context.JSON(http.StatusOK, map[string]interface{}{"token": ss})
	} else {
		return context.String(http.StatusBadRequest, "Code refused")
	}
}

