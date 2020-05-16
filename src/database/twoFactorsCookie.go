package database

import (
	"github.com/go-playground/validator/v10"
	"time"
)

//context.RealIP()

// This should always be handled server side (no direct modification from client)

type TwoFactorsCookie struct {
	BaseModel
	Uuid		string `json:"uuid" validate:"uuid4" gorm:"type:varchar(36)"`
	IpAddr		string `json:"ip_addr" validate:"ipv4" gorm:"type:varchar(12)"`
	UserAgent	string `json:"user_agent" validate:"alphanumunicode" gorm:"type:varchar(200)"`
	Expires		time.Time `json:"expires"`
	LastUsed	time.Time `json:"last_used"`
	UserID 		uint
}

func (t TwoFactorsCookie) Validate() error {
	validate = validator.New()

	return validate.Struct(&t)
}

func (t *TwoFactorsCookie) Update() error {
	return GetDB().Save(&t).Error
}

func (t *TwoFactorsCookie) Create() error {
	return GetDB().Create(&t).Error
}

func (t *TwoFactorsCookie) Delete() error {
	return GetDB().Delete(&t).Error
}
