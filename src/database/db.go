package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var mu sync.Mutex
var initialized uint32 = 0
var instance *gorm.DB

type BaseModel struct{
	ID        uint `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at"`
}

func GetDB() *gorm.DB {
	if atomic.LoadUint32(&initialized) == 1 {
		return instance
	}
	mu.Lock()
	defer mu.Unlock()

	if initialized == 0 {
		instance = Connect()
		instance.LogMode(true)
		instance.Set("gorm:auto_preload", true)

		atomic.StoreUint32(&initialized, 1)
	}

	return instance
}


func Connect() *gorm.DB {
	var uri = "user=" + os.Getenv("DIARY_DB_USER") +
		" host=diary-postgres password=" + os.Getenv("DIARY_DB_PWD") + " sslmode=disable"

	db, err := gorm.Open("postgres", uri)
	if err != nil {
		log.Fatalln("failed to connect database", err)
	}
	log.Println("Connected to database")
	return db
}

func RunMigration() {
	instance.AutoMigrate(&User{})
	instance.AutoMigrate(&Entry{})
	instance.AutoMigrate(&Label{})

	//	instance.Create(&models.User{Email: "toto@address.com"})
	//	db.Create(&models.User{Email: "tzata@tata.com", Password:"azer"})
}
