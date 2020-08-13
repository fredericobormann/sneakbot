package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Draw struct {
	gorm.Model
	GroupchatID   int64
	Participant   Participant
	ParticipantID uint
	Time          time.Time
}
