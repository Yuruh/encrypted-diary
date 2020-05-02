package database

import (
	"github.com/go-playground/validator/v10"
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestLabel_Validate(t *testing.T) {
	assert := asserthelper.New(t)
	var label Label = Label{
		BaseModel:    BaseModel{},
		PartialLabel: PartialLabel{
			Name: "toto",
			Color: "validate me",
		},
		UserID:       0,
	}

	err := label.Validate()
	assert.NotNil(err)
	assert.IsType(validator.ValidationErrors{}, err)

	casted := err.(validator.ValidationErrors)

	assert.Equal("Validation failed on field: 'Color'. Expected to respect rule: 'hexcolor'. Got value: 'validate me'.\n",
		BuildValidationErrorMsg(casted))
}

func TestLabel_Create(t *testing.T) {
	assert := asserthelper.New(t)
	var label Label = Label{
		PartialLabel: PartialLabel{
			Name: "toto",
			Color:   "color",
		},
	}
	err := label.Create()
	assert.Nil(err)

	var foundLabel Label
	GetDB().Where("id = ?", label.ID).First(&foundLabel)
	assert.Equal(foundLabel.Name, "toto")

	err = label.Create()
	assert.NotNil(err)
}

func TestLabel_Update(t *testing.T) {
	assert := asserthelper.New(t)
	var label Label = Label{
		PartialLabel: PartialLabel{
			Name: "toto",
			Color:   "color",
		},
	}

	GetDB().Create(&label)
	label.Name = "The updated name"
	err := label.Update()
	assert.Nil(err)

	var foundLabel Label
	GetDB().Where("id = ?", label.ID).First(&foundLabel)
	assert.Equal(foundLabel.Name, "The updated name")
}

func TestLabel_Delete(t *testing.T) {
	assert := asserthelper.New(t)
	var label Label = Label{
		PartialLabel: PartialLabel{
			Name: "toto",
			Color:   "color",
		},
	}

	GetDB().Create(&label)
	err := label.Delete()
	assert.Nil(err)

	var foundLabel Label
	result := GetDB().Where("id = ?", label.ID).First(&foundLabel)
	assert.Equal(true, result.RecordNotFound())
}