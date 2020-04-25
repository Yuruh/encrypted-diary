package database

type TwoFactorsCookie struct {
	BaseModel
	Uuid		string `json:"uuid" validate:"uuid4" gorm:"type:varchar(36)"`
	IpAddr		string `json:"ip_addr" validate:"ipv4" gorm:"type:varchar(12)"`
	UserAgent	string `json:"user_agent" validate:"alphanumunicode" gorm:"type:varchar(200)"`
}