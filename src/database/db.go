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
	DeletedAt *time.Time `sql:"index" json:"-"`
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
//		instance.Set("gorm:auto_preload", true)

		atomic.StoreUint32(&initialized, 1)
		RunMigration()
	}

	return instance
}

func Connect() *gorm.DB {
	log.Println("Connecting to database...")
	var uri = "user=" + os.Getenv("DIARY_DB_USER") +
		" password=" + os.Getenv("DIARY_DB_PWD") +
		" host=" + os.Getenv("DB_HOST") +
		" sslmode=disable"

	if os.Getenv("DIARY_DB_NAME") != "" {
		uri += " dbname=" + os.Getenv("DIARY_DB_NAME")
	}

	db, err := gorm.Open("postgres", uri)
	if err != nil {
		log.Fatalln("failed to connect database", err)
	}
	log.Println("Connected to database")
	return db
}

/*
	Auto Migrate seems to fail to create foreign keys, hence creation of a many to many relation for entries / label failed

	https://github.com/jinzhu/gorm/issues/450#issuecomment-487958084
 */
func RunMigration() {
	instance.Exec("CREATE EXTENSION fuzzystrmatch")

	instance.AutoMigrate(&User{})
	instance.AutoMigrate(&Entry{})
	instance.AutoMigrate(&Label{})
	instance.AutoMigrate(&TwoFactorsCookie{})
}
