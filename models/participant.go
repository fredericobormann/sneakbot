package models

import "github.com/jinzhu/gorm"

type Participant struct {
	gorm.Model
	GroupchatId int64
	UserId      int
}
