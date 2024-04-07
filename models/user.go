package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Firstname string     `json:"firstname" validate:"required,min=3,max=30"`
	Lastname  string     `json:"lastname" validate:"required,min=3,max=30"`
	Username  string     `json:"username" validate:"omitempty,min3,max=30"`
	Email     string     `json:"email" validate:"required,email"`
	Password  string     `json:"password" validate:"required,min=7"`
	CreatedAt time.Time  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (u *User) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

func (u *User) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(u)
}

func (u *User) Validate() error {
	v := validator.New()
	return v.Struct(u)
}
