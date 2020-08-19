package database

import (
	"errors"
	"github.com/fredericobormann/sneakbot/models"
	"github.com/fredericobormann/sneakbot/texts"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"math/rand"
	"sort"
)

var t = true
var f = false

type Datastore struct {
	DB *gorm.DB
}

func New() *Datastore {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.Participant{})

	return &Datastore{
		DB: db,
	}
}

func (store *Datastore) AddOrUpdateGroup(groupChatId int64, latestPollId int) tgbotapi.Chattable {
	invalidatedPoll := store.invalidateOldPoll(groupChatId)
	var group models.Group
	t := true
	store.DB.Where(models.Group{GroupchatId: groupChatId}).Assign(models.Group{LatestPollId: latestPollId, Activated: &t}).FirstOrCreate(&group)
	return invalidatedPoll
}

func (store *Datastore) invalidateOldPoll(groupChatId int64) tgbotapi.Chattable {
	var checkGroup models.Group
	store.DB.Where(models.Group{GroupchatId: groupChatId}).First(&checkGroup)
	if checkGroup.GroupchatId != 0 {
		editPoll := tgbotapi.NewEditMessageText(groupChatId, checkGroup.LatestPollId, texts.Expired_message)
		editPoll.ReplyMarkup = nil
		return editPoll
	}
	return nil
}

func (store *Datastore) DeactivateGroup(groupChatId int64) {
	f := false
	var group models.Group
	store.DB.Where(models.Group{GroupchatId: groupChatId}).First(&group)
	if group.Activated != nil {
		group.Activated = &f
		store.DB.Save(&group)
	}
}

func (store *Datastore) AddParticipant(groupChatId int64, userId int, firstName string, lastName string) bool {
	var participant models.Participant
	var formerParticipants []models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId, Active: &t}).Find(&formerParticipants)
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId, FirstName: firstName, LastName: lastName}).Assign(models.Participant{Active: &t}).FirstOrCreate(&participant)
	return len(formerParticipants) == 0
}

func (store *Datastore) RemoveParticipant(groupChatId int64, userId int) bool {
	var formerParticipants []models.Participant
	var deletedParticipant models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId, Active: &t}).Find(&formerParticipants)
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId}).Assign(models.Participant{Active: &f}).FirstOrCreate(&deletedParticipant)
	return len(formerParticipants) > 0
}

func (store *Datastore) ResetGroup(groupChatId int64) {
	store.DB.Model(models.Participant{}).Where(models.Participant{GroupchatId: groupChatId}).Updates(models.Participant{Active: &f})
}

func (store *Datastore) GetParticipants(groupChatId int64) []models.Participant {
	var participants []models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId, Active: &t}).Find(&participants)
	return participants
}

func (store *Datastore) GetNRandomParticipants(groupChatId int64, numberOfPeople int) ([]models.Participant, error) {
	var participants []models.Participant
	participants = store.GetParticipants(groupChatId)

	if len(participants) < numberOfPeople {
		return []models.Participant{}, errors.New("Not enough participants")
	}
	rand.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})

	sort.SliceStable(participants, func(i, j int) bool {
		return participants[i].Counter < participants[j].Counter
	})

	for _, p := range participants[:numberOfPeople] {
		store.DB.Where("id = ?", p.ID).Assign(models.Participant{Counter: p.Counter + 1}).FirstOrCreate(&models.Participant{})
	}

	return participants[:numberOfPeople], nil
}

func (store *Datastore) GetAllParticipantsWithoutName() []models.Participant {
	var participants []models.Participant
	store.DB.Where("first_name IS NULL AND last_name IS NULL").Find(&participants)
	return participants
}

func (store *Datastore) SetNameOfParticipant(userID int, firstName string, lastName string) {
	var participant models.Participant
	store.DB.Model(&participant).Where("user_id = ?", userID).Update("first_name", firstName).Update("last_name", lastName)
}

func (store *Datastore) GetAllGroups() []models.Group {
	var groups []models.Group
	t := true
	store.DB.Where(models.Group{Activated: &t}).Find(&groups)
	return groups
}
