package models

import "github.com/jinzhu/gorm"

type Participant struct {
	gorm.Model
	GroupchatId int64
	UserId      int
	FirstName   string
	LastName    string
}

func (p Participant) GetFullName() string {
	return p.FirstName + " " + p.LastName
}
