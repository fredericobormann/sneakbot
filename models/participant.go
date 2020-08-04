package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

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

func (p Participant) GetMarkup() string {
	return fmt.Sprintf("[%v](tg://user?id=%d)", p.GetFullName(), p.UserId)
}
