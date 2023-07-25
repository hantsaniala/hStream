package hStream

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	db  *gorm.DB
	err error
)

func init() {
	var (
		host     = GetEnv("DB_HOST")
		port     = GetEnv("DB_PORT")
		user     = GetEnv("DB_USER")
		dbname   = GetEnv("DB_NAME")
		password = GetEnv("DB_PASSWORD")
	)

	conn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host,
		port,
		user,
		dbname,
		password,
	)

	db, err = gorm.Open("postgres", conn)
	db.AutoMigrate(Video{})

	if err != nil {
		log.Fatal(err)
	}
}
