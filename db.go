package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDb() *gorm.DB {
	db := connectToDb()
	migrateDb(db)
	return db
}

func connectToDb() *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=0.0.0.0 port=5432 user=postgres password=pass dbname=kindle_db",
	}), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect to database")
	}
	return db
}

func migrateDb(db *gorm.DB) {
	db.AutoMigrate(&NotedItem{}, &Word{})
}
