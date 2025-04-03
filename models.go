package main

import "time"

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique"`
	Password string
}

// File model to store uploaded file details
type File struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index"`    // Foreign key to associate with users
	Filename    string    `gorm:"not null"` // Original file name
	FilePath    string    `gorm:"not null"` // Path where the file is stored
	FileType    string    `gorm:"not null"` // MIME type (e.g., image/png, application/pdf)
	Size        int64     `gorm:"not null"` // File size in bytes
	Description string    // Optional file description
	TinyURL     string    `gorm:"unique"` // Short URL for sharing
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Title       string    `gorm:"not null"`
	UniqueID    string    `gorm:"not null"`
}

// Auto-migrate the model
func MigrateDB() {
	err := DB.AutoMigrate(&File{})
	if err != nil {
		panic("❌ Database migration failed: " + err.Error())
	}
}
