package database

import (
	"github.com/go-playground/validator/v10"
)

type User struct {
	BaseModel
	Email       string  `gorm:"type:varchar(100);unique_index" json:"email" validate:"email,required"`
	Password	string  `gorm:"not null" json:"-"`
	Entries		[]Entry	`json:"-"`
	Labels		[]Label `json:"-"`

	// Stored encrypted
	OTPSecret	string 	`json:"-"`
	HasRegisteredOTP bool `json:"has_registered_otp"`
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

	return validate.Struct(&user)
}

func (user *User) Delete() error {
	return GetDB().Delete(&user).Error
}