package database

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type User struct {
	BaseModel
	Email       string  `gorm:"type:varchar(100);unique_index" json:"email" validate:"email,required"`
	Password	string  `gorm:"not null" json:"-"`
	Entries		[]Entry	`json:"entries"`
	//	ApiAccess	ApiAccess `gorm:"foreignKey:UserID" json:",omitempty"`
}

func (user *User) Create() error {
	db := GetDB().Create(&user)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}
	return nil
}

func (user *User) Update() error {
	db := GetDB().Save(&user)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}

	return nil
}

func (user User) Validate() error {
	validate = validator.New()

	err := validate.Struct(&user)
	if err != nil {

		errorMsg := BuildValidationErrorMsg(err.(validator.ValidationErrors))

		//todo figure out how to send this message through the error
		fmt.Println(errorMsg)

		return err
	}
	return nil
}