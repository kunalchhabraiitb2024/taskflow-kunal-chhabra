package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never serialised to JSON
	CreatedAt time.Time `json:"created_at"`
}

// PublicUser is what we return in API responses — no password field.
type PublicUser struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func (u *User) ToPublic() PublicUser {
	return PublicUser{ID: u.ID, Name: u.Name, Email: u.Email}
}
