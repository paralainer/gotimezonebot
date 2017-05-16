package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
)

type TgBot struct {
	Location *MongoTimezonesService
	Api      *tgbotapi.BotAPI
	Weather  GetWeather
	chats map[int64]*Chat
}

func StartBot(token string, tzService *MongoTimezonesService, weather GetWeather) {
	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot := &TgBot{
		Api:      botApi,
		Location: tzService,
		Weather:  weather,
		chats: make(map[int64]*Chat, 10),
	}

	bot.startBot()
}

func (bot *TgBot) startBot() {

	log.Printf("Authorized on account %s", bot.Api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.Api.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message != nil {
			chat, ok := bot.chats[update.Message.Chat.ID]
			if !ok {
				chat = NewChat(update.Message.Chat.ID, *bot)
				bot.chats[update.Message.Chat.ID] = chat
			}

			go chat.ProcessMessage(update.Message)
		}
	}
}

