package api

import (
	"github.com/Yuruh/encrypted-diary/src"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func RunHttpServer()  {
	// Echo instance
	app := echo.New()
	app.HideBanner = true

	// Middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())

//	app.POST("/login", login)
//	app.POST("/register", register)

//	unprotectedPaths := [2]string{"/login", "/register"}

	// According to https://echo.labstack.com/middleware, "Middleware registered using Echo#Use() is only executed for paths which are registered after Echo#Use() has been called."
	// But it doesn't behave that way so for now we'll skip specific routes
/*	app.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims: &TokenClaims{},
		SigningKey: []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		SigningMethod: "HS256",
		ContextKey: "token",
		Skipper: func(context echo.Context) bool {
			if utils.ContainsString(unprotectedPaths[:], context.Path()) {
				return true
			}
			return false
		},
		SuccessHandler: func(context echo.Context) {
			context.Set("user", context.Get("token").(*jwt.Token).Claims.(*TokenClaims).User)
		},
	}))*/

	// Routes
	app.GET("/", hello)
	app.GET("/openapi.yml", sendApiSpec)

	app.GET("/entries", src.GetEntries)
	app.POST("/entries", src.AddEntry)

	// Start server
	app.Logger.Fatal(app.Start(":8080"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func sendApiSpec(c echo.Context) error {
	return c.File("openapi.yml")
}

func (c TokenClaims) Valid() error {
	return c.StandardClaims.Valid()
}

type TokenClaims struct {
	User database.User `json:"user"`
	jwt.StandardClaims
}

/*func login(context echo.Context) error {
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody models.User

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		log.Println(err.Error())
		return context.String(http.StatusBadRequest, "")
	}
	var user models.User
	database.GetDB().Where("email = ?", parsedBody.Email).First(&user)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(parsedBody.Password))

	if err != nil {
		println("User not found: " + err.Error())
		return context.String(http.StatusNotFound, "User not found")
	} else {
		user.Password = ""
		claims := &TokenClaims{
			user,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Unix() + int64(time.Hour * 24),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, _ := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

		return context.JSON(http.StatusOK, map[string]interface{}{"token": ss, "user": user})
	}
}

func register(context echo.Context) error {
	body, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var parsedBody models.User

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		log.Println(err.Error())
		return context.String(http.StatusBadRequest, "")
	}

	// todo validate input

	hash, err := bcrypt.GenerateFromPassword([]byte(parsedBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return context.String(http.StatusBadRequest, "Could not hash password")
	}

	clonedDb := database.GetDB().Create(&models.User{Email: parsedBody.Email, Password: string(hash)})

	if clonedDb.Error != nil {
		println(clonedDb.Error.Error())
		return context.NoContent(http.StatusBadRequest)
	}

	return context.String(http.StatusOK, "")
}
*/