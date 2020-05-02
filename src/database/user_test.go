package database

import (
	"github.com/go-playground/validator/v10"
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	assert := asserthelper.New(t)
	var user User = User{
		Email: "email",
		Password: "toto",
	}

	err := user.Validate()
	assert.NotNil(err)
	assert.IsType(validator.ValidationErrors{}, err)

	casted := err.(validator.ValidationErrors)

	assert.Equal("Validation failed on field: 'Email'. Expected to respect rule: 'email'. Got value: 'email'.\n",
		BuildValidationErrorMsg(casted))
}

func TestUser_Create(t *testing.T) {
	assert := asserthelper.New(t)
	GetDB().Unscoped().Delete(User{})
	var user User = User{
		Email: "email",
		Password: "toto",
	}
	err := user.Create()
	assert.Nil(err)

	var foundUser User
	GetDB().Where("id = ?", user.ID).First(&foundUser)
	assert.Equal(foundUser.Email, "email")

	err = user.Create()
	assert.NotNil(err)
}

func TestUser_Update(t *testing.T) {
	assert := asserthelper.New(t)
	GetDB().Unscoped().Delete(User{})
	var user User = User{
		Email: "email",
		Password: "toto",
	}

	var user2 User = User{
		Email: "other_email",
		Password: "toto",
	}

	GetDB().Create(&user)
	GetDB().Create(&user2)
	user.Email = "The updated email"
	err := user.Update()
	assert.Nil(err)

	var foundUser User
	GetDB().Where("id = ?", user.ID).First(&foundUser)
	assert.Equal(foundUser.Email, "The updated email")

	user2.Email = "The updated email"
	err = user2.Update()
	assert.NotNil(err)
}

func TestUser_Delete(t *testing.T) {
	assert := asserthelper.New(t)
	GetDB().Unscoped().Delete(User{})
	var user User = User{
		Email: "email",
		Password: "toto",
	}

	GetDB().Create(&user)
	err := user.Delete()
	assert.Nil(err)

	var foundUser User
	result := GetDB().Where("id = ?", user.ID).First(&foundUser)
	assert.Equal(true, result.RecordNotFound())
}