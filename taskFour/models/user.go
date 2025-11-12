package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Posts     []Post    `gorm:"foreignKey:UserID" json:"-"`
	Comments  []Comment `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HashPassword 加密密码
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// BeforeCreate GORM钩子，在创建前加密密码
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.HashPassword()
}
