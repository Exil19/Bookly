package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserName string  `json:"username" gorm:"unique"`
	Email    string  `json:"email" gorm:"uniqueIndex;not null"`
	Password string  `json:"password" gorm:"not null"`
	Profile  Profile `json:"profile"`
	Books    []Book  `json:"books"`
}

type Profile struct {
	gorm.Model
	UserID uint   `json:"user_id" gorm:"uniqueIndex"`
	Avatar string `json:"avatar" gorm:"type:text"`
	Bio    string `json:"bio" gorm:"type:text"`
}
