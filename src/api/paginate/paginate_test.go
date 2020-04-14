package paginate

import (
	"github.com/Yuruh/encrypted-diary/src/database"
	asserthelper "github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestGetPaginationResults(t *testing.T) {
	assert := asserthelper.New(t)
	database.GetDB().Unscoped().Delete(database.Entry{})

	for i := 0; i < 13; i++ {
		entry := database.Entry{
			PartialEntry: database.PartialEntry{
				Content: strconv.Itoa(i),
				Title:   "Entry " + strconv.Itoa(i),
			},
		}
		_ = database.Insert(&entry)
	}


	pagination, err := GetPaginationResults("entries", 3, 2)
	assert.Nil(err)
	assert.Equal((uint)(2), pagination.Page)
	assert.Equal((uint)(3), pagination.Limit)
	assert.Equal((uint)(3), pagination.NextPage)
	assert.Equal((uint)(1), pagination.PrevPage)
	assert.Equal((uint)(13), pagination.TotalMatches)
	assert.Equal(true, pagination.HasNextPage)
	assert.Equal(true, pagination.HasPrevPage)
	assert.Equal((uint)(5), pagination.TotalPages)

	pagination, err = GetPaginationResults("unexist", 4, 2)
	assert.NotNil(err)

	pagination, err = GetPaginationResults("entries", 0, 2)
	assert.NotNil(err)
}