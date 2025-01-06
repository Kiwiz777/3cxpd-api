package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type StorageClient struct {
	DB *gorm.DB
}

func NewStorageClient() (*StorageClient, error) {

	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USERNAME")
	password := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	log.Println(password)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, username, password, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	db.AutoMigrate(&User{}, &Token{}, &Contact{}, &Action{}, &SystemKey{})

	client := &StorageClient{
		DB: db,
	}
	SeedDB(db)
	return client, nil
}

func GenerateToken(length int) string {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		panic(err)
	}
	randomString := base64.URLEncoding.EncodeToString(buffer)[:length]

	return randomString
}