package main

import (
	"github.com/go-co-op/gocron"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"sneakbot/database"
	"sneakbot/texts"
	"strings"
	"time"
)

type Config struct {
	Webhook struct {
		Url      string `yaml:"url"`
		ApiToken string `yaml:"apikey"`
	} `yaml:"webhook"`
}

var participationReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(texts.Button_yes, "yes_participant"),
		tgbotapi.NewInlineKeyboardButtonData(texts.Button_no, "no_participant"),
	),
)

var cfg Config
var bot *tgbotapi.BotAPI

func init() {
	f, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
}

func handleMessage(update tgbotapi.Update) {
	msgtext := update.Message.Text
	var err error
	if strings.HasPrefix(msgtext, "/") {
		if strings.HasPrefix(msgtext, "/start") {
			err = handleCommandStart(update)
		} else if strings.HasPrefix(msgtext, "/reset") {
			err = handleCommandReset(update)
		} else if strings.HasPrefix(msgtext, "/draw") {
			err = handleCommandDraw(update)
		} else if strings.HasPrefix(msgtext, "/stop") {
			err = handleCommandStop(update)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func handleCommandStart(update tgbotapi.Update) error {
	err := sendPoll(update, texts.Start_message+"\n\n"+getParticipantsText(update.Message.Chat.ID))
	return err
}

func handleCommandReset(update tgbotapi.Update) error {
	database.ResetGroup(update.Message.Chat.ID)
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, texts.Reset_message)
	_, errSend := bot.Send(answer)
	if errSend != nil {
		log.Fatal(errSend)
	}
	err := sendPoll(update, texts.Start_message)
	return err
}

func handleCommandDraw(update tgbotapi.Update) error {
	return sendNewRandomParticipants(update.Message.Chat.ID)
}

func sendNewRandomParticipants(groupChatId int64) error {
	randomParticipants, errRandom := database.GetNRandomParticipants(groupChatId, 2)
	if errRandom != nil {
		msg := tgbotapi.NewMessage(groupChatId, texts.Not_enough_participants)
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Fatal(errSend)
		}
		return nil
	}
	var participantsText string
	for _, p := range randomParticipants {
		participantsText += getFullNameOfUser(groupChatId, p.UserId) + "\n"
	}
	answer := tgbotapi.NewMessage(groupChatId, texts.Random_participants_drawn+participantsText)
	_, err := bot.Send(answer)
	return err
}

func sendAllNewRandomParticipants() {
	groups := database.GetAllGroups()
	for _, g := range groups {
		err := sendNewRandomParticipants(g.GroupchatId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleCommandStop(update tgbotapi.Update) error {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, texts.Stop_message)
	_, err := bot.Send(answer)
	if err == nil {
		database.DeactivateGroup(update.Message.Chat.ID)
	}
	return err
}

func sendPoll(update tgbotapi.Update, msgText string) error {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	answer.ReplyMarkup = participationReplyMarkup
	msg, err := bot.Send(answer)
	if err == nil {
		invalidatedPoll := database.AddOrUpdateGroup(update.Message.Chat.ID, msg.MessageID)
		if invalidatedPoll != nil {
			_, err := bot.Send(invalidatedPoll)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return err
}

func handleCallbackQuery(update tgbotapi.Update) {
	if update.CallbackQuery.Data == "yes_participant" {
		handleNewParticipant(update)
	} else if update.CallbackQuery.Data == "no_participant" {
		handleDeleteParticipant(update)
	}
}

func updatePollResult(update tgbotapi.Update) {
	participantsText := getParticipantsText(update.CallbackQuery.Message.Chat.ID)
	editedPollMessage := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		texts.Start_message+"\n\n"+
			participantsText,
	)
	editedPollMessage.ReplyMarkup = &participationReplyMarkup
	_, err := bot.Send(editedPollMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func getParticipantsText(groupChatId int64) string {
	participants := database.GetParticipants(groupChatId)
	participantsText := ""
	for _, p := range participants {
		participantsText += getFullNameOfUser(groupChatId, p.UserId) + "\n"
	}
	return participantsText
}

func getFullNameOfUser(groupChatId int64, userId int) string {
	chatmember, err := bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: groupChatId,
		UserID: userId,
	})
	if err != nil {
		return "Unknown User"
	}
	return chatmember.User.FirstName + " " + chatmember.User.LastName
}

func handleNewParticipant(update tgbotapi.Update) {
	changed := database.AddParticipant(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.ID)
	_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, texts.New_participant_message))
	if err != nil {
		log.Println(err)
	}
	if changed {
		updatePollResult(update)
	}
}

func handleDeleteParticipant(update tgbotapi.Update) {
	changed := database.RemoveParticipant(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.From.ID)
	_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, texts.Delete_participant_message))
	if err != nil {
		log.Println(err)
	}
	if changed {
		updatePollResult(update)
	}
}

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI(cfg.Webhook.ApiToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(cfg.Webhook.Url + bot.Token))
	if err != nil {
		log.Fatal(err)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8443", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	scheduler := gocron.NewScheduler(time.UTC)
	_, errScheduler := scheduler.Every(1).Wednesday().At("12:00:00").Do(sendAllNewRandomParticipants)
	if errScheduler != nil {
		log.Println(errScheduler)
	}
	scheduler.StartAsync()

	for update := range updates {
		log.Printf("%+v\n", update)
		if update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) {
			chatMember, _ := bot.GetChatMember(
				tgbotapi.ChatConfigWithUser{
					ChatID: update.Message.Chat.ID,
					UserID: update.Message.From.ID,
				},
			)
			if chatMember.IsCreator() || chatMember.IsAdministrator() {
				handleMessage(update)
			}
		} else if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, texts.No_groupchat)
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(update)
		}
	}
}
