package database

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) {
	log.Println("Seeding database")

	// Seed Users
	Users := []User{
		{
			ID:         uuid.New(),
			Username:   "admin",
			Email:      "admin@rashwan.com",
			Password:   "d033e22ae348aeb5660fc2140aec35850c4da997", // Note: Use hashed password in production admin
			FirstName:  "Admin",
			LastName:   "Admin",
		},
	}
	for _, user := range Users {
		existingUser := User{}
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
			log.Printf("User '%s' already exists, skipping", user.Username)
			continue
		}

		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Error seeding Users: %v", err)
		}
	}
	//seed system key
	SystemKeys := []SystemKey{
		{
			ID:  uuid.New(),
			Key: "3cxpdRashwan",
		},
	}
	for _, key := range SystemKeys {
		existingKey := SystemKey{}
		if err := db.Where("key = ?", key.Key).First(&existingKey).Error; err == nil {
			log.Printf("Key '%s' already exists, skipping", key.Key)
			continue
		}

		if err := db.Create(&key).Error; err != nil {
			log.Fatalf("Error seeding SystemKeys: %v", err)
		}
	}

	log.Println("Database seeded successfully")
}
