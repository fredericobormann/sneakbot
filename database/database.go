package database

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"sneakbot/texts"
)

var db *gorm.DB

type Group struct {
	gorm.Model
	GroupchatId  int64
	LatestPollId int
	Activated    *bool `gorm:"default:true"`
}

type Participant struct {
	gorm.Model
	GroupchatId int64
	UserId      int
}

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Group{})
	db.AutoMigrate(&Participant{})
}

func AddOrUpdateGroup(groupChatId int64, latestPollId int) tgbotapi.Chattable {
	invalidatedPoll := invalidateOldPoll(groupChatId)
	var group Group
	t := true
	db.Where(Group{GroupchatId: groupChatId}).Assign(Group{LatestPollId: latestPollId, Activated: &t}).FirstOrCreate(&group)
	return invalidatedPoll
}

func invalidateOldPoll(groupChatId int64) tgbotapi.Chattable {
	var checkGroup Group
	db.Where(Group{GroupchatId: groupChatId}).First(&checkGroup)
	fmt.Println(&checkGroup)
	if checkGroup.GroupchatId != 0 {
		editPoll := tgbotapi.NewEditMessageText(groupChatId, checkGroup.LatestPollId, texts.Expired_message)
		editPoll.ReplyMarkup = nil
		return editPoll
	}
	return nil
}

func DeactivateGroup(groupChatId int64) {
	f := false
	var group Group
	db.Where(Group{GroupchatId: groupChatId}).First(&group)
	if group.Activated != nil {
		group.Activated = &f
		db.Save(&group)
	}
}

func AddParticipant(groupChatId int64, userId int) {
	var participant Participant
	db.Where(Participant{GroupchatId: groupChatId, UserId: userId}).FirstOrCreate(&participant)
}

func RemoveParticipant(groupChatId int64, userId int) {
	db.Where(Participant{GroupchatId: groupChatId, UserId: userId}).Delete(Participant{})
}

func ResetGroup(groupChatId int64) {
	db.Where(Participant{GroupchatId: groupChatId}).Delete(Participant{})
}
