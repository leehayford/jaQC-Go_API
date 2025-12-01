package api

import (
	"jaQC-Go-API/utils"
)

type User struct {
	utils.Meta     	`gorm:"embedded"`
	Password string `gorm:"type:varchar(100);not null" json:"-"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Role     string `json:"role"`
}
func (User) TableName() string { return "users" }

/* TRANSPORT OBJECT */
type UserRegistrationInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" validate:"required,min=8"`
}

type UserLoginInput struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}