package paginate

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
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
	page = 1

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

// take a *db as param with the where cause already applied
func GetPaginationResults(table string, limit uint, page uint, db *gorm.DB) (Pagination, error) {
	var pagination Pagination
	if limit == 0 {
		return Pagination{}, errors.New("trying to divide by 0")
	}

	// Gets total matches
	err := db.Select("COUNT(*) as total_matches").Table(table).Scan(&pagination).Error
	if err != nil {
		sentry.CaptureException(err)
		return Pagination{}, err
	}
	pagination.TotalPages = pagination.TotalMatches / limit
	if pagination.TotalPages % limit != 0 || limit > pagination.TotalMatches {
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

	return pagination, nil
}