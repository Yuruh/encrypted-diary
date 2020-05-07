package api

import (
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/Yuruh/encrypted-diary/src/helpers"
	"github.com/dgrijalva/jwt-go"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func AuthMiddleware() echo.MiddlewareFunc {
	unprotectedPaths := [4]string{"/login", "/register", "/openapi.yml", "/auth/two-factors/otp/authenticate"}

	return middleware.JWTWithConfig(middleware.JWTConfig{
		Claims: &TokenClaims{},
		SigningKey: []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		SigningMethod: "HS512",
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

func BuildRateLimiterConf() *limiter.Limiter {
	// create a limiter with expirable token buckets
	// create a 3 request/second limiter and
	// every token bucket in it will expire 10 minutes after it was initially set.
	lmt := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute * 10})

	// Places to look for IP Addr
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})

	return lmt
}

// Middleware to limit the number of request an IP can make during a time window
func RateLimiterMiddleware(limiter *limiter.Limiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return echo.HandlerFunc(func(c echo.Context) error {
			fmt.Println(c.Request().Header)
			httpError := tollbooth.LimitByRequest(limiter, c.Response(), c.Request())
			if httpError != nil {
				fmt.Println("http errooor")
				return c.String(httpError.StatusCode, httpError.Message)
			}
			return next(c)
		})
	}
}

// Middleware to recover from panic and send infos to sentry

func RecoverMiddleware() echo.MiddlewareFunc {
	// Defaults
	config := middleware.DefaultRecoverConfig
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					hub := sentry.CurrentHub()
					hub.Recover(r)
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						c.Logger().Printf("[PANIC RECOVER] %v %s\n", strings.Replace(err.Error(), `\n`, "\n", -1), strings.Replace(string(stack[:length]), `\n`, "\n", -1))
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
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
	app.Use(RecoverMiddleware())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.BodyLimit("1G"))
	app.Use(RateLimiterMiddleware(BuildRateLimiterConf()))

	app.POST("/login", Login, RequireBody)
	app.POST("/register", Register, RequireBody)


	// According to https://echo.labstack.com/middleware, "Middleware registered using Echo#Use() is only executed for paths which are registered after Echo#Use() has been called."
	// But it doesn't behave that way so for now we'll skip specific routes
	app.Use(AuthMiddleware())
	// Routes
	app.GET("/openapi.yml", SendApiSpec)

	app.GET("/me", GetMe)

	app.GET("/entries", GetEntries)
	app.GET("/entries/:id", GetEntry)
	app.POST("/entries", AddEntry, RequireBody)
	app.PUT("/entries/:id", EditEntry, RequireBody)
	app.DELETE("/entries/:id", DeleteEntry)

	app.GET("/labels", GetLabels)
	app.POST("/labels", AddLabel, RequireBody)
	app.PUT("/labels/:id", EditLabel, RequireBody, middleware.BodyLimit("150K"))
	app.DELETE("/labels/:id", DeleteLabel)

	app.POST("/auth/two-factors/otp/register", RequestGoogleAuthenticatorQRCode)
	app.GET("/auth/two-factors/otp/token", RequestTwoFactorsToken)
	app.POST("/auth/two-factors/otp/authenticate", ValidateOTPCode)
}

// TODO
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

