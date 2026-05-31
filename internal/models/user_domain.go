package models

import (
	"database/sql"
	"time"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusInActive  UserStatus = "inactive"
)

type User struct {
	Id                  string       `json:"id" gorm:"column:id;primaryKey"`
	Name                string       `json:"name" gorm:"column:name"`
	Username            string       `json:"username" gorm:"column:username"`
	Email               string       `json:"email" gorm:"column:email"`
	Password            string       `json:"-" gorm:"column:password"`
	Status              UserStatus   `json:"status" gorm:"column:status;default:active"`
	IsVerified          bool         `json:"is_verified" gorm:"column:is_verified"`
	ForcePasswordChange bool         `json:"force_password_change" gorm:"column:force_password_change"`
	OnboardingCompleted bool         `json:"onboarding_completed" gorm:"column:onboarding_completed"`
	LastLogin           time.Time    `json:"last_login" gorm:"column:last_login"`
	CreatedAt           time.Time    `json:"created_at" gorm:"column:created_at"`
	UpdatedAt           time.Time    `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt           sql.NullTime `json:"deleted_at" gorm:"column:deleted_at"`
}

func (u *User) TableName() string {
	return "users"
}
