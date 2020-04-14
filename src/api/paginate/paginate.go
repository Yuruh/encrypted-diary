package paginate

import (
	"errors"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"strconv"
)

type Pagination struct {
	Limit uint	`json:"limit"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
	Page uint `json:"page"`
	NextPage uint `json:"next_page"`
	PrevPage uint `json:"prev_page"`
	TotalPages uint `json:"total_pages"`
	TotalMatches uint `json:"total_matches"`
}


func GetPaginationParams(defaultLimit int, context echo.Context) (limit int, page int, offset int, err error) {
	limit = defaultLimit
	offset = 0
	page = 0

	if context.QueryParam("limit") != "" {
		limit, err = strconv.Atoi(context.QueryParam("limit"))
		if err != nil {
			return 1, 1, 0, err
		}
	}

	if context.QueryParam("page") != "" {
		page, err = strconv.Atoi(context.QueryParam("page"))
		if err != nil {
			return 1, 1, 0, err
		}
		if page >= 1 {
			offset = limit * (page - 1)
		}
	}
	if limit <= 0 {
		return 1, 1, 0, errors.New("limit must be > 0")
	}
	return limit, page, offset, nil
}

// todo add where
func GetPaginationResults(table string, limit uint, page uint) (Pagination, error) {
	var pagination Pagination
	if limit == 0 {
		return Pagination{}, errors.New("trying to divide by 0")
	}

	// Gets total matches
	db := database.GetDB().Raw("SELECT COUNT(*) as total_matches FROM " + table).Scan(&pagination)

	pagination.TotalPages = pagination.TotalMatches / limit
	if pagination.TotalPages % limit != 0 {
		pagination.TotalPages++
	}
	pagination.Page = page
	pagination.HasPrevPage = page > 1
	pagination.HasNextPage = pagination.TotalPages > page

	if pagination.HasPrevPage {
		pagination.PrevPage = page - 1
	}
	if pagination.HasNextPage {
		pagination.NextPage = page + 1
	}
	pagination.Limit = limit

	// This method should be preferred as it prevents sqi injection, but there's an error
	// SELECT COUNT('id') as total_matches FROM "entries" fails because of the quotes. The request does work from pg admin though
	// db = database.GetDB().Table(table).Select("COUNT(?) as total_matches", "id").Scan(&test)
	if db.Error != nil {
		return Pagination{}, db.Error
	}
	return pagination, nil
}