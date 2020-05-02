package paginate

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetPaginationParams(t *testing.T) {
	assert := asserthelper.New(t)
	e := echo.New()
	request, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	context := e.NewContext(request, recorder)


	limit, page, offset, err := GetPaginationParams(21, context)
	assert.Nil(err)
	assert.Equal(21, limit)
	assert.Equal(1, page)
	assert.Equal(0, offset)

	context.QueryParams().Set("limit", "3")
	context.QueryParams().Set("page", "1")

	limit, page, offset, err = GetPaginationParams(1, context)
	assert.Nil(err)
	assert.Equal(3, limit)
	assert.Equal(1, page)
	assert.Equal(0, offset)

	context.QueryParams().Set("page", "2")
	limit, page, offset, err = GetPaginationParams(1, context)
	assert.Nil(err)
	assert.Equal(3, limit)
	assert.Equal(2, page)
	assert.Equal(3, offset)

	context.QueryParams().Set("page", "3")
	limit, page, offset, err = GetPaginationParams(1, context)
	assert.Nil(err)
	assert.Equal(3, limit)
	assert.Equal(3, page)
	assert.Equal(6, offset)


	context.QueryParams().Set("limit", "-1")
	_, _, _, err = GetPaginationParams(1, context)
	assert.NotNil(err)

	context.QueryParams().Set("limit", "patate")
	_, _, _, err = GetPaginationParams(1, context)
	assert.NotNil(err)

	context.QueryParams().Set("limit", "10")
	context.QueryParams().Set("page", "patate")
	_, _, _, err = GetPaginationParams(1, context)
	assert.NotNil(err)

}

func TestGetPaginationResults(t *testing.T) {
	assert := asserthelper.New(t)
	database.GetDB().Unscoped().Delete(database.Entry{})

	for i := 0; i < 13; i++ {
		entry := database.Entry{
			PartialEntry: database.PartialEntry{
				Content: "i love pagination",
				Title:   "Entry " + strconv.Itoa(i),
			},
		}
		_ = database.Insert(&entry)
	}

	pagination, err := GetPaginationResults("entries", 3, 2, database.GetDB().Where("content = ?", "i love pagination"))
	assert.Nil(err)
	assert.Equal((uint)(2), pagination.Page)
	assert.Equal((uint)(3), pagination.Limit)
	assert.Equal((uint)(3), pagination.NextPage)
	assert.Equal((uint)(1), pagination.PrevPage)
	assert.Equal((uint)(13), pagination.TotalMatches)
	assert.Equal(true, pagination.HasNextPage)
	assert.Equal(true, pagination.HasPrevPage)
	assert.Equal((uint)(5), pagination.TotalPages)

	pagination, err = GetPaginationResults("unexist", 4, 2, database.GetDB())
	assert.NotNil(err)

	pagination, err = GetPaginationResults("entries", 0, 2, database.GetDB())
	assert.NotNil(err)


	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "allo",
			Title:   "Paginate me",
		},
	}
	_ = database.Insert(&entry)

	pagination, err = GetPaginationResults("entries", 5, 1, database.GetDB().Where("id = ?", entry.ID))
	assert.Nil(err)
	assert.Equal((uint)(1), pagination.Page)
	assert.Equal((uint)(5), pagination.Limit)
	assert.Equal((uint)(0), pagination.NextPage)
	assert.Equal((uint)(0), pagination.PrevPage)
	assert.Equal((uint)(1), pagination.TotalMatches)
	assert.Equal(false, pagination.HasNextPage)
	assert.Equal(false, pagination.HasPrevPage)
	assert.Equal((uint)(1), pagination.TotalPages)
}