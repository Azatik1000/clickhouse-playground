package models

import "github.com/jinzhu/gorm"

type Result string

type Run struct {
	gorm.Model
	Query  Query `gorm:"EMBEDDED"`
	Result Result
}
