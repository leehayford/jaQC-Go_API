package utils

import (
	// "github.com/google/uuid" // go get github.com/google/uuid
)

type Meta struct {
	// ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"` // "github.com/google/uuid"
	// ID int64 `gorm:"unique; primaryKey"` // POSTGRESS
	ID int64 `gorm:"autoIncrement" json:"id"` // SQLITE
	
	CreatedAt int64     `gorm:"autoCreateTime:milli" json:"created_at"`
	CreatedBy int64   `json:"created_by"` // UserID
	
	UpdatedAt int64     `gorm:"autoUpdateTime:milli" json:"updated_at"`
	UpdatedBy int64  `json:"updated_by"` // UserID
	
	DeletedAt int64 `json:"deleted_at"`// Time:milli
}