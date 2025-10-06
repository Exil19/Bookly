package models

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Name    string `json:"name" gorm:"type:text;not null;unique"`
	Author  string `json:"author" gorm:"type:text;not null"`
	Image   string `json:"image" gorm:"type:text;not null"`
	UserID  uint   `json:"user_id"`
	Creator User   `json:"creator" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
