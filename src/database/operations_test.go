package database

import (
	"github.com/go-playground/validator/v10"
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestInsert(t *testing.T) {
	assert := asserthelper.New(t)

	var entry Entry = Entry{
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
	}

	err := Insert(&entry)
	assert.NotNil(err)
	assert.IsType(validator.ValidationErrors{}, err)

	entry.Title = "valid title"

	err = Insert(&entry)
	assert.Nil(err)
	err = Insert(&entry)
	assert.NotNil(err)
}

func TestUpdate(t *testing.T) {
	assert := asserthelper.New(t)

	var entry Entry = Entry{
		PartialEntry: PartialEntry{
			Content: "the content",
			Title:   "a",
		},
	}
	GetDB().Create(&entry)
	entry.Content = "The updated content"
	err := Update(&entry)
	assert.NotNil(err)
	assert.IsType(validator.ValidationErrors{}, err)

	entry.Title = "valid title"
	err = Update(&entry)
	assert.Nil(err)

	var foundEntry Entry
	GetDB().Where("id = ?", entry.ID).First(&foundEntry)
	assert.Equal(foundEntry.Content, "The updated content")
	assert.Equal(foundEntry.Title, "valid title")

}