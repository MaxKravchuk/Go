package models

import "gorm.io/gorm"

type Urls struct {
	gorm.Model
	id  int `gorm:"primaryKey`
	Url string
}
