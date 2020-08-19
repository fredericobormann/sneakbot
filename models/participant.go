package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

type Participant struct {
	gorm.Model
	GroupchatId int64
	UserId      int
	FirstName   string
	LastName    string
	Active      *bool `gorm:"default:true"`
	Counter     int   `gorm:"default:0"`
}

func (p Participant) GetFullName() string {
	fullName := p.FirstName + " " + p.LastName
	trimmedName := strings.TrimSpace(fullName)
	escapedName := strings.ReplaceAll(trimmedName, ".", "\\.")
	return escapedName
}

func (p Participant) GetMarkup() string {
	return fmt.Sprintf("[%v](tg://user?id=%d)", p.GetFullName(), p.UserId)
}
