package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func (h *Handler) AddNamesOfUsersToDB() {
	allParticipantsWithoutName := h.Datastore.GetAllParticipantsWithoutName()
	for _, p := range allParticipantsWithoutName {
		chatmember, err := h.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: p.GroupchatId,
			UserID: p.UserId,
		})
		if err != nil {
			log.Printf("Migration failed for user id %d", p.UserId)
			return
		}
		h.Datastore.SetNameOfParticipant(p.UserId, chatmember.User.FirstName, chatmember.User.LastName)
	}
}
