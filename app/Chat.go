package app

import (
	"gopkg.in/telegram-bot-api.v4"
	"sort"
	"strings"
	"time"
	"log"
	"fmt"
)

type ChatState struct {
	ChatID   int64
	State    string
	Location *Location
}

type Chat struct {
	State *ChatState
	bot   *TgBot
}

func NewChat(chatId int64, bot *TgBot) *Chat {
	return &Chat{
		State: &ChatState{
			ChatID: chatId,
			State: no_state,
		},
		bot:    bot,
	}
}
func RestoreChat(state *ChatState, bot *TgBot) *Chat {
	return &Chat{
		State: state,
		bot:   bot,
	}
}

const (
	no_state                    = "no_state"
	adding_location_name        = "adding_location_name"
	adding_location_coordinates = "adding_location_coordinates"
	confirming_location         = "confirming_location"
	removing_location           = "removing_location"
)

func (chat *Chat) ProcessMessage(message *tgbotapi.Message) {
	if message.IsCommand() {
		chat.processCommand(message)
	} else if message.Location != nil {
		location := message.Location
		chat.sendLocationTimeAndWeather(location.Latitude, location.Longitude)
	} else if chat.State.State != no_state {
		chat.processConversation(message)
	}
}

func (chat *Chat) sendText(text string) {
	newMessage := tgbotapi.NewMessage(chat.State.ChatID, text)
	newMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	chat.bot.Api.Send(newMessage)
}

func (chat *Chat) sendWithReply(text string, replyMarkup interface{}) {
	newMessage := tgbotapi.NewMessage(chat.State.ChatID, text)
	newMessage.ReplyMarkup = replyMarkup
	chat.bot.Api.Send(newMessage)
}

func (chat *Chat) processCommand(message *tgbotapi.Message) {
	chat.State.State = no_state
	arguments := message.CommandArguments()
	switch message.Command() {
	case "tztime":
		if arguments != "" {
			chat.sendLocationTimeAndWeatherByAlias(arguments)
		} else {
			chat.sendTime(message)
		}
		break
	case "addtz":
		if arguments != "" {
			chat.setLocationName(arguments)
		} else {
			chat.startAddLocation(message)
		}
		break
	case "rmtz":
		if arguments != "" {
			chat.removeLocationByAlias(arguments)
		} else {
			chat.startRemoveLocation(message)
		}
		break

	}
}

func (chat *Chat) processConversation(message *tgbotapi.Message) {
	switch chat.State.State {
	case removing_location:
		chat.removeLocationByAlias(message.Text)
		break
	case adding_location_name:
		chat.setLocationName(message.Text)
		break
	case confirming_location:
		chat.confirmLocation(message.Text)
		break
	case adding_location_coordinates:
		chat.setLocationCoordinates(message.Location)

	}
}

func (chat *Chat) confirmLocation(answer string) {
	if answer == "Yes" {
		chat.State.State = no_state
		newLocation := chat.State.Location
		chat.bot.Location.AddLocation(chat.State.ChatID, *newLocation)
		chat.sendText("Added")
	} else {
		chat.State.State = adding_location_coordinates
		chat.sendText("Send me a location of that place")
	}

}

func (chat *Chat) setLocationCoordinates(location *tgbotapi.Location) {
	if location != nil {
		chat.State.State = no_state
		newLocation := chat.State.Location
		newLocation.Coordinates = formatCoordinates(location.Latitude, location.Longitude)
		chat.bot.Location.AddLocation(chat.State.ChatID, *newLocation)
		chat.sendText("Added")
	} else {
		chat.State.State = adding_location_coordinates
		chat.sendText("Please send location")
	}
}

func (chat *Chat) setLocationName(locName string) {
	if strings.Trim(locName, " ") != "" {
		if chat.State.Location == nil {
			chat.State.Location = &Location{}
		}
		loc := chat.State.Location
		loc.Alias = locName
		geoInfo, err := GetGeoInfo(locName)
		if err != nil {
			chat.State.State = adding_location_coordinates
			chat.sendText("Send me a location of that place")
		} else {
			chat.sendText(geoInfo.LocationDisplayName)
			chat.bot.Api.Send(tgbotapi.NewLocation(chat.State.ChatID, geoInfo.Lat, geoInfo.Lon))
			chat.sendWithReply("Is that correct location?", tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Yes"),
					tgbotapi.NewKeyboardButton("No"),
				)))
			loc.Coordinates = formatCoordinates(geoInfo.Lat, geoInfo.Lon)
			chat.State.State = confirming_location
		}
	}
}

func (chat *Chat) startRemoveLocation(message *tgbotapi.Message) {
	locations := chat.bot.Location.GetChatLocations(message.Chat.ID)
	buttons := [][]tgbotapi.KeyboardButton{}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].Alias < locations[j].Alias
	})
	for _, loc := range locations {
		buttons = append(buttons, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(loc.Alias)))
	}

	chat.sendWithReply("Send me location name to delete", tgbotapi.NewReplyKeyboard(buttons...))
	chat.State.State = removing_location
}

func (chat *Chat) removeLocationByAlias(alias string) {
	chat.State.State = no_state
	alias = strings.Trim(alias, " ")
	deleted := chat.bot.Location.RemoveTimezone(chat.State.ChatID, alias)
	var text string
	if deleted {
		text = "Deleted"
	} else {
		text = "Not Found"
	}

	chat.sendWithReply(text, tgbotapi.NewRemoveKeyboard(false))
}

func (chat *Chat) startAddLocation(message *tgbotapi.Message) {
	chat.State.State = adding_location_name
	chat.State.Location = &Location{}
	chat.sendText("Send me location name")
}

func (chat *Chat) sendTime(message *tgbotapi.Message) {
	locations := chat.bot.Location.GetChatLocations(message.Chat.ID)
	filteredLocations := make([]Location, len(locations))
	for _, loc := range locations {
		if loc.Coordinates != "" && loc.Alias != "" {
			filteredLocations = append(filteredLocations, loc)
		}
	}
	chat.sendText(chat.convertTzToString(locations))
}

func (chat *Chat) convertTzToString(locations []Location) string {
	result := make([]string, len(locations))
	currentTime := time.Now()
	weatherChan := make(chan WeatherResult)
	for _, location := range locations {
		go chat.bot.Weather(location, weatherChan)
	}

	for range locations {
		weatherResult := <-weatherChan
		if weatherResult.Error != nil {
			log.Println(weatherResult.Error)
			result = append(result, "Error occurred: "+weatherResult.Error.Error())
		} else {
			weather := weatherResult.Weather
			formattedTzTime := formatTzTime(weather.Timezone, currentTime)
			result = append(result, weather.Location.Alias+": "+formattedTzTime+weather.Conditions)
		}

	}

	sort.Strings(result)

	return strings.Join(result, "\n")
}

func (chat * Chat) sendLocationTimeAndWeatherByAlias(alias string) {
	geoInfo, err := GetGeoInfo(alias)
	if err != nil {
		chat.sendText("Can't find such place")
		return
	}
	chat.sendText(geoInfo.LocationDisplayName)
	chat.bot.Api.Send(tgbotapi.NewLocation(chat.State.ChatID, geoInfo.Lat, geoInfo.Lon))
	chat.sendLocationTimeAndWeather(geoInfo.Lat, geoInfo.Lon)
}

func (chat *Chat) sendLocationTimeAndWeather(lat float64, lng float64) {
	weatherChan := make(chan WeatherResult)
	go chat.bot.Weather(Location{Coordinates: formatCoordinates(lat, lng)}, weatherChan)
	weatherResult := <-weatherChan

	if weatherResult.Error != nil {
		log.Println(weatherResult.Error)
		chat.sendText("Error occurred: " + weatherResult.Error.Error())
	} else {
		weather := weatherResult.Weather
		formattedTzTime := formatTzTime(weather.Timezone, time.Now())
		chat.sendText(formattedTzTime + weather.Conditions)
	}
}

func formatCoordinates(lat float64, lng float64) string {
	return fmt.Sprintf("%.6f",lat) + "," + fmt.Sprintf("%.6f", lng)
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
