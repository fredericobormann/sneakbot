package handler

import (
	"fmt"
	"github.com/fredericobormann/sneakbot/database"
	"github.com/fredericobormann/sneakbot/texts"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

type Handler struct {
	Datastore *database.Datastore
	Bot       *tgbotapi.BotAPI
}

var participationReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(texts.Button_yes, "yes_participant"),
		tgbotapi.NewInlineKeyboardButtonData(texts.Button_no, "no_participant"),
	),
)

func New(db *database.Datastore, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		Datastore: db,
		Bot:       bot,
	}
}

func (handler *Handler) HandleMessage(update tgbotapi.Update) {
	msgtext := update.Message.Text
	var err error
	if strings.HasPrefix(msgtext, "/") {
		if strings.HasPrefix(msgtext, "/start") {
			err = handler.handleCommandStart(update)
		} else if strings.HasPrefix(msgtext, "/reset") {
			err = handler.handleCommandReset(update)
		} else if strings.HasPrefix(msgtext, "/draw") {
			err = handler.handleCommandDraw(update)
		} else if strings.HasPrefix(msgtext, "/stop") {
			err = handler.handleCommandStop(update)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (handler *Handler) handleCommandStart(update tgbotapi.Update) error {
	err := handler.sendPoll(update, texts.Start_message+"\n\n"+handler.getParticipantsText(update.Message.Chat.ID))
	return err
}

func (handler *Handler) handleCommandReset(update tgbotapi.Update) error {
	handler.Datastore.ResetGroup(update.Message.Chat.ID)
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, texts.Reset_message)
	_, errSend := handler.Bot.Send(answer)
	if errSend != nil {
		log.Fatal(errSend)
	}
	err := handler.sendPoll(update, texts.Start_message)
	return err
}

func (handler *Handler) handleCommandDraw(update tgbotapi.Update) error {
	return handler.sendNewRandomParticipants(update.Message.Chat.ID)
}

func (handler *Handler) sendNewRandomParticipants(groupChatId int64) error {
	randomParticipants, errRandom := handler.Datastore.GetNRandomParticipants(groupChatId, 2)
	if errRandom != nil {
		msg := tgbotapi.NewMessage(groupChatId, texts.Not_enough_participants)
		_, errSend := handler.Bot.Send(msg)
		if errSend != nil {
			log.Fatal(errSend)
		}
		return nil
	}
	var participantsText string
	for _, p := range randomParticipants {
		participantsText += p.GetMarkup() + "\n"
	}
	answer := tgbotapi.NewMessage(groupChatId, texts.Random_participants_drawn+participantsText)
	answer.ParseMode = "MarkdownV2"
	_, err := handler.Bot.Send(answer)
	return err
}

func (handler *Handler) SendAllNewRandomParticipants() {
	groups := handler.Datastore.GetAllGroups()
	for _, g := range groups {
		err := handler.sendNewRandomParticipants(g.GroupchatId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (handler *Handler) handleCommandStop(update tgbotapi.Update) error {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, texts.Stop_message)
	_, err := handler.Bot.Send(answer)
	if err == nil {
		handler.Datastore.DeactivateGroup(update.Message.Chat.ID)
	}
	return err
}

func (handler *Handler) sendPoll(update tgbotapi.Update, msgText string) error {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	answer.ParseMode = "MarkdownV2"
	answer.ReplyMarkup = participationReplyMarkup
	msg, err := handler.Bot.Send(answer)
	if err == nil {
		invalidatedPoll := handler.Datastore.AddOrUpdateGroup(update.Message.Chat.ID, msg.MessageID)
		if invalidatedPoll != nil {
			_, err := handler.Bot.Send(invalidatedPoll)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return err
}

func (handler *Handler) HandleCallbackQuery(update tgbotapi.Update) {
	if update.CallbackQuery.Data == "yes_participant" {
		handler.handleNewParticipant(update)
	} else if update.CallbackQuery.Data == "no_participant" {
		handler.handleDeleteParticipant(update)
	}
}

func (handler *Handler) updatePollResult(update tgbotapi.Update) {
	participantsText := handler.getParticipantsText(update.CallbackQuery.Message.Chat.ID)
	editedPollMessage := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		texts.Start_message+"\n\n"+
			participantsText,
	)
	editedPollMessage.ParseMode = "MarkdownV2"
	editedPollMessage.ReplyMarkup = &participationReplyMarkup
	_, err := handler.Bot.Send(editedPollMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func (handler *Handler) getParticipantsText(groupChatId int64) string {
	participants := handler.Datastore.GetParticipants(groupChatId)
	var participantsText string
	if len(participants) == 1 {
		participantsText = fmt.Sprintf(texts.Participants_message_one, len(participants))
	} else {
		participantsText = fmt.Sprintf(texts.Participants_message_many, len(participants))
	}
	for _, p := range participants {
		participantsText += p.GetMarkup() + "\n"
	}
	return participantsText
}

func (handler *Handler) handleNewParticipant(update tgbotapi.Update) {
	changed := handler.Datastore.AddParticipant(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.ID, update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName)
	_, err := handler.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, texts.New_participant_message))
	if err != nil {
		log.Println(err)
	}
	if changed {
		handler.updatePollResult(update)
	}
}

func (handler *Handler) handleDeleteParticipant(update tgbotapi.Update) {
	changed := handler.Datastore.RemoveParticipant(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.ID)
	_, err := handler.Bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, texts.Delete_participant_message))
	if err != nil {
		log.Println(err)
	}
	if changed {
		handler.updatePollResult(update)
	}
}
