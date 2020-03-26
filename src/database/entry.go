package database

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

/*
	The user modifiable part of Entry
*/
type PartialEntry struct {
	Content		string `json:"content" gorm:"type:varchar"`
	Title		string `json:"title" gorm:"type:varchar" validate:"required,min=3"`
}

/*
	The remaining data, generated by the server
*/
type Entry struct {
	BaseModel
	PartialEntry
	UserID uint
}

func (entry *Entry) Create() error {
	db := GetDB().Create(&entry)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}
	return nil
}

func (entry *Entry) Update() error {
	db := GetDB().Save(&entry)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}

	return nil
}

func (entry Entry) Validate() error {
	validate = validator.New()

	return validate.Struct(&entry)
/*	if err != nil {

		errorMsg := BuildValidationErrorMsg(err.(validator.ValidationErrors))

		fmt.Println(errorMsg)

		return err
	}
	return nil*/
}

/*type ValidationError struct {
	msg string
	err error
}

func (e *ValidationError) Error() string {
	return e.msg
}

func (e *ValidationError) Unwrap() error {
	return e.err
}*/