package database

import (
	"errors"
	"github.com/fredericobormann/sneakbot/models"
	"github.com/fredericobormann/sneakbot/texts"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"math/rand"
	"time"
)

type Datastore struct {
	DB *gorm.DB
}

type StatisticResult struct {
	GroupchatID   int64
	ParticipantID uint
	Participant   models.Participant
	Total         uint
}

func New() *Datastore {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.Participant{})
	db.AutoMigrate(&models.Draw{})

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
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId}).Find(&formerParticipants)
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId, FirstName: firstName, LastName: lastName}).FirstOrCreate(&participant)
	return len(formerParticipants) == 0
}

func (store *Datastore) RemoveParticipant(groupChatId int64, userId int) bool {
	var formerParticipants []models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId}).Find(&formerParticipants)
	store.DB.Where(models.Participant{GroupchatId: groupChatId, UserId: userId}).Delete(models.Participant{})
	return len(formerParticipants) > 0
}

func (store *Datastore) ResetGroup(groupChatId int64) {
	store.DB.Where(models.Participant{GroupchatId: groupChatId}).Delete(models.Participant{})
}

func (store *Datastore) GetParticipants(groupChatId int64) []models.Participant {
	var participants []models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId}).Find(&participants)
	return participants
}

func (store *Datastore) GetNRandomParticipants(groupChatId int64, numberOfPeople int) ([]models.Participant, error) {
	var participants []models.Participant
	store.DB.Where(models.Participant{GroupchatId: groupChatId}).Find(&participants)
	if len(participants) < numberOfPeople {
		return []models.Participant{}, errors.New("Not enough participants")
	}
	rand.Shuffle(len(participants), func(i, j int) {
		participants[i], participants[j] = participants[j], participants[i]
	})
	for _, p := range participants[:numberOfPeople] {
		store.DB.Create(&models.Draw{Participant: p, GroupchatID: groupChatId, Time: time.Now()})
	}

	store.GetStatisticByGroupId(-125444678)

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

func (store *Datastore) GetStatisticByGroupId(groupchatID int64) []StatisticResult {
	var statisticResults []StatisticResult
	store.DB.Table("draws").Select("groupchat_id, participant_id, count(id) as total").Where("deleted_at IS NULL AND groupchat_id = ?", groupchatID).Group("groupchat_id, participant_id").Scan(&statisticResults)
	for i := range statisticResults {
		var participant models.Participant
		store.DB.Where("id = ?", statisticResults[i].ParticipantID).First(&participant)
		statisticResults[i].Participant = participant
	}
	return statisticResults
}
