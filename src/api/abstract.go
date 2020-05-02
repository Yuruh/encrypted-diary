package api

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

/*
	Abstract implementation for an http call DELETE /resource/:id
	Expect the resource to be associated to the user with the foreign key user_id
*/
func DeleteAbstract(context echo.Context, m database.Model) error {
	var user database.User = context.Get("user").(database.User)

	id, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		return context.String(http.StatusBadRequest, "Bad route parameter")
	}

	result := database.GetDB().
		Where("ID = ?", id).
		Where("user_id = ?", user.ID).
		First(m)
	if result.RecordNotFound() {
		return context.NoContent(http.StatusNotFound)
	} else if result.Error != nil {
		return InternalError(context, result.Error)
	}
	err = m.Delete()
	if err != nil {
		return InternalError(context, result.Error)
	}
	return context.NoContent(http.StatusOK)
}
