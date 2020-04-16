package api

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"time"
)

func AuthMiddleware() echo.MiddlewareFunc {
	unprotectedPaths := [3]string{"/login", "/register", "/openapi.yml"}

	return middleware.JWTWithConfig(middleware.JWTConfig{
		Claims: &TokenClaims{},
		SigningKey: []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		SigningMethod: "HS256",
		ContextKey: "token",
		Skipper: func(context echo.Context) bool {
			if helpers.ContainsString(unprotectedPaths[:], context.Path()) {
				return true
			}
			return false
		},
		SuccessHandler: func(context echo.Context) {
			context.Set("user", context.Get("token").(*jwt.Token).Claims.(*TokenClaims).User)
		},
	})
}

func RequireBody(next echo.HandlerFunc) echo.HandlerFunc {
	return func (c echo.Context) error {
		if c.Request().Body == nil {
			return c.String(http.StatusBadRequest, "Body required")
		}
		return next(c)
	}
}

func DeclareRoutes(app *echo.Echo) {
	// Middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())
	app.Use(middleware.BodyLimit("1G"))

	app.POST("/login", Login, RequireBody)
	app.POST("/register", Register, RequireBody)


	// According to https://echo.labstack.com/middleware, "Middleware registered using Echo#Use() is only executed for paths which are registered after Echo#Use() has been called."
	// But it doesn't behave that way so for now we'll skip specific routes
	app.Use(AuthMiddleware())

	// Routes
	app.GET("/", hello)
	app.GET("/openapi.yml", SendApiSpec)

	app.GET("/entries", GetEntries)
	app.GET("/entries/:id", GetEntry)
	app.POST("/entries", AddEntry, RequireBody)
	app.PUT("/entries/:id", EditEntry, RequireBody)
	app.DELETE("/entries/:id", DeleteEntry)

	app.GET("/labels", GetLabels)
	app.POST("/labels", AddLabel, RequireBody)

	// todo test bodylimit middleware
	app.PUT("/labels/:id", EditLabel, RequireBody, middleware.BodyLimit("6M"))
	app.DELETE("/labels/:id", DeleteLabel)
	
}

func TokenToRemainingDuration() time.Duration {
	return time.Hour * 1
}

func RunHttpServer()  {
	// Echo instance
	app := echo.New()
	app.HideBanner = true

	DeclareRoutes(app)
	// Start server
	app.Logger.Fatal(app.Start(":8080"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func SendApiSpec(c echo.Context) error {
	return c.File("openapi.yml")
}

func (c TokenClaims) Valid() error {
	return c.StandardClaims.Valid()
}

type TokenClaims struct {
	User database.User `json:"user"`
	jwt.StandardClaims
}

