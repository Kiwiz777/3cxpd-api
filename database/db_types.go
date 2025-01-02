package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`

}

type Token struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expires_at"`
}

type Contact struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Name	  string    `json:"name"`
	Number	  string    `json:"number"`
	CFStatus  string    `json:"cf_status"`
}
type Action struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	ContactID uuid.UUID `json:"contact_id"`
	Caller  string	`json:"caller"`
	CallTime string	`json:"call_time"`
	Notes string	`json:"notes"`
}

type SystemKey struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Key       string    `json:"key"`
}