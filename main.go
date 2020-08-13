package main

import (
	"github.com/fredericobormann/sneakbot/database"
	"github.com/fredericobormann/sneakbot/handler"
	"github.com/fredericobormann/sneakbot/texts"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Webhook struct {
		Url      string `yaml:"url"`
		ApiToken string `yaml:"apikey"`
	} `yaml:"webhook"`
}

func main() {
	go StartStatisticServer()

	cfg, err := readConfig()
	if err != nil {
		log.Fatal("Reading config unsuccessful.")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Webhook.ApiToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	db := database.New()
	h := handler.New(db, bot)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(cfg.Webhook.Url + bot.Token))
	if err != nil {
		log.Fatal(err)
	}

	h.AddNamesOfUsersToDB()

	updates := bot.ListenForWebhook("/" + bot.Token)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8443", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	scheduler := gocron.NewScheduler(time.UTC)
	_, errScheduler := scheduler.Every(1).Wednesday().At("12:00:00").Do(h.SendAllNewRandomParticipants)
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
				h.HandleMessage(update)
			}
		} else if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, texts.No_groupchat)
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else if update.CallbackQuery != nil {
			h.HandleCallbackQuery(update)
		}
	}
}

func readConfig() (*Config, error) {
	f, err := os.Open("config.yml")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	cfg := &Config{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func StartStatisticServer() {
	router := gin.Default()
	datastore := database.New()
	statisticHandler := handler.NewStatisticHandler(datastore)
	router.GET("/:groupchat_id", statisticHandler.HandleStatisticRequestByGroup)

	err := router.Run(":8090")
	if err != nil {
		log.Fatal("Could not start statistic server", err)
	}
}
