package api

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSendApiSpec(t *testing.T) {
	err := os.Chdir("../..")
	if err != nil {
		t.Fatal(err.Error())
	}
	e := echo.New()
	request, err := http.NewRequest("GET", "/openapi.yml", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = SendApiSpec(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}
	if len(recorder.Body.String()) < 400 {
		t.Error("Spec content insufficient")
	}
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("No token", caseNoToken)
	t.Run("Bad token", caseBadToken)
	t.Run("Bad signing", caseBadTokenSigning)
	t.Run("Unprotected", caseUnprotectedRoute)
	t.Run("Expired", caseExpiredToken)
	t.Run("Valid", caseValidToken)
}

func caseNoToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)

	if err.(*echo.HTTPError).Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, err.(*echo.HTTPError).Code)
	}
}

func caseBadToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Add("Authorization", "Bearer toto")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)

	if err.(*echo.HTTPError).Code != http.StatusUnauthorized {
		t.Errorf("Bad status, expected %v, got %v", http.StatusUnauthorized, err.(*echo.HTTPError).Code)
	}
}

func caseBadTokenSigning(t *testing.T) {
	claims := &TokenClaims{
		database.User{
			BaseModel: database.BaseModel{ID: 321},
		},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + int64(time.Hour * 24),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte("BAD KEY"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	req.Header.Add("Authorization", "Bearer " + ss)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)

	if err.(*echo.HTTPError).Code != http.StatusUnauthorized {
		t.Errorf("Bad status, expected %v, got %v (%v)",
			http.StatusUnauthorized, err.(*echo.HTTPError).Code, err.(*echo.HTTPError).Message)
	}
}

func caseUnprotectedRoute(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/login")

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)

	httpErr := err.(*echo.HTTPError)
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusNotFound, httpErr.Code, httpErr.Message)
	}
}

func caseExpiredToken(t *testing.T) {
	claims := &TokenClaims{
		database.User{
			BaseModel: database.BaseModel{ID:321},
		},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() - 10000,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	req.Header.Add("Authorization", "Bearer " + ss)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)

	httpErr := err.(*echo.HTTPError)
	if err.(*echo.HTTPError).Code != http.StatusUnauthorized {
		t.Errorf("Bad status, expected %v, got %v", http.StatusUnauthorized, err.(*echo.HTTPError).Code)
	}
	if httpErr.Message != "invalid or expired jwt" {
		t.Errorf("Bad message, expected %v, got %v", "invalid or expired jwt", httpErr.Message)
	}
}

func caseValidToken(t *testing.T) {
	claims := &TokenClaims{
		database.User{
			BaseModel: database.BaseModel{ID:321},
		},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + int64(time.Hour * 24),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

	e := echo.New()
	e.GET("/", func(ctx echo.Context) error {return ctx.NoContent(http.StatusOK)})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	req.Header.Add("Authorization", "Bearer " + ss)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var auth = AuthMiddleware()(echo.NotFoundHandler)
	err := auth(c)
	httpErr := err.(*echo.HTTPError)
	// all's well, but the route hasn't been found
	// But success handler is untested though
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusNotFound, httpErr.Code, httpErr)
	}
}

func TestDeclareRoutes(t *testing.T) {
	e := echo.New()
	assert := asserthelper.New(t)
	DeclareRoutes(e)
	assert.Equal(13, len(e.Routes()))
}

func TestRecoverMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	funcRecover := RecoverMiddleware()(echo.HandlerFunc(func(c echo.Context) error {
		panic("test")
	}))

	err := funcRecover(c)
	asserthelper.Nil(t, err)
	asserthelper.Equal(t, http.StatusInternalServerError, rec.Code)
}