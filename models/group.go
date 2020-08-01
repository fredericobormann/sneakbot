package models

import "github.com/jinzhu/gorm"

type Group struct {
	gorm.Model
	GroupchatId  int64
	LatestPollId int
	Activated    *bool `gorm:"default:true"`
}
