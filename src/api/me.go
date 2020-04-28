package api

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetMe(c echo.Context) error {
	var user database.User = c.Get("user").(database.User)

	return c.JSON(http.StatusOK, map[string]interface{}{"user": user})
}