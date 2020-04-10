package database

import (
	"github.com/go-playground/validator/v10"
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestEntry_Validate(t *testing.T) {
	assert := asserthelper.New(t)
	var entry Entry = Entry{
		BaseModel:    BaseModel{},
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
		UserID:       0,
	}

	err := entry.Validate()
	assert.NotNil(err)
	assert.IsType(validator.ValidationErrors{}, err)

	casted := err.(validator.ValidationErrors)

	assert.Equal("Validation failed on field: 'Title'. Expected to respect rule: 'min 3'. Got value: 'a'.\n",
		BuildValidationErrorMsg(casted))
}