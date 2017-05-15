package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"time"
	"strings"
	"sort"
)

type tgBot struct {
	TzService *MongoTimezonesService
	Api       *tgbotapi.BotAPI
	Weather   GetWeather
}

func StartBot(token string, tzService *MongoTimezonesService, weather GetWeather) {
	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot := &tgBot{
		Api:       botApi,
		TzService: tzService,
		Weather:   weather,
	}

	bot.startBot()
}

func (bot *tgBot) startBot() {

	log.Printf("Authorized on account %s", bot.Api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.Api.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}
		go bot.onCommand(update.Message)
	}
}

func (bot *tgBot) onCommand(message *tgbotapi.Message) {
	command := message.Command()
	switch command {
	case "tztime":
		bot.sendTime(message)
		break
	case "addtz":
		bot.addTz(message)
	}
}

func (bot *tgBot) addTz(message *tgbotapi.Message) {
	arguments := message.CommandArguments()
	parts := strings.SplitN(arguments, " ", 3)
	if len(parts) != 3 {
		bot.Api.Send(tgbotapi.NewMessage(message.Chat.ID, "Usage: /addtz"))
		return
	}

	chatTz := &Location{
		Alias:               parts[1],
	}

	bot.TzService.AddTimezone(message.Chat.ID, chatTz)
}

func (bot *tgBot) sendTime(message *tgbotapi.Message) {
	bot.Api.Send(
		tgbotapi.NewMessage(
			message.Chat.ID,
			bot.convertTzToString(
				bot.TzService.GetChatTimezones(message.Chat.ID))))
}

func (bot *tgBot) convertTzToString(locations []Location) string {
	result := make([]string, len(locations))
	currentTime := time.Now()
	weatherChan := make(chan Weather)
	for _, location := range locations {
		go bot.Weather(location, weatherChan)
	}

	for range locations {
		weather := <-weatherChan
		formattedTzTime := formatTzTime(weather.Timezone, currentTime)
		result = append(result, weather.Location.Alias+": "+formattedTzTime+weather.Conditions)

	}

	sort.Strings(result)

	return strings.Join(result, "\n")
}

func formatTzTime(tz string, currentTime time.Time) (string) {
	location, err := time.LoadLocation(tz)
	if err != nil {
		log.Panic(err)
	}
	t := currentTime.In(location)
	formattedTzTime := t.Format("Jan _2, 2006 15:04")
	return formattedTzTime
}

