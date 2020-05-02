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

func TestEntry_Create(t *testing.T) {
	assert := asserthelper.New(t)
	var entry Entry = Entry{
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
	}
	err := entry.Create()
	assert.Nil(err)

	var foundEntry Entry
	GetDB().Where("id = ?", entry.ID).First(&foundEntry)
	assert.Equal(foundEntry.Content, "the content")

	err = entry.Create()
	assert.NotNil(err)
}

func TestEntry_Update(t *testing.T) {
	assert := asserthelper.New(t)
	var entry Entry = Entry{
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
	}

	GetDB().Create(&entry)
	entry.Content = "The updated content"
	err := entry.Update()
	assert.Nil(err)

	var foundEntry Entry
	GetDB().Where("id = ?", entry.ID).First(&foundEntry)
	assert.Equal(foundEntry.Content, "The updated content")

/*
	Not sure how to make it fail, as gorm Save() will insert if doesn't exist
	err = entry.Update()

	assert.NotNil(err)*/
}

func TestEntry_Delete(t *testing.T) {
	assert := asserthelper.New(t)
	var entry Entry = Entry{
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
	}

	GetDB().Create(&entry)
	err := entry.Delete()
	assert.Nil(err)

	var foundEntry Entry
	result := GetDB().Where("id = ?", entry.ID).First(&foundEntry)
	assert.Equal(true, result.RecordNotFound())

/*
entry.ID = 12345678
   	err = entry.Delete()
   	assert.NotNil(err)

This does not fail, only affects 0 rows
 */
}