package database

import "github.com/go-playground/validator/v10"

// The user modifiable part
type PartialLabel struct {
	Name string `json:"name" validate:"alphanumunicode,max=100"`
	Color string `json:"color" validate:"hexcolor"`
}

type Label struct {
	BaseModel
	PartialLabel
	UserID uint `json:"user_id"`
	AvatarUrl string `json:"avatar_url"`
//	Entries		[]Entry `json:"entries" gorm:"many2many:entry_labels;"`
}

func (label *Label) Create() error {
	db := GetDB().Create(&label)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}
	return nil
}

func (label *Label) Update() error {
	db := GetDB().Save(&label)
	if db.Error != nil {
		println(db.Error.Error())
		return db.Error
	}

	return nil
}

func (label Label) Validate() error {
	// We could generate a color here if none is given
	// If so, we should maybe separate validate into two differents method, e.g. IsValid() and FixValidation()
	validate = validator.New()

	return validate.Struct(&label)
}