package app

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/aws/aws-lambda-go/lambda"
	"context"
	"os"
	"github.com/aws/aws-lambda-go/events"
	"encoding/json"
)

type TgBot struct {
	Location *MongoLocationsService
	Api      *tgbotapi.BotAPI
	Weather  GetWeather
	chats map[int64]*Chat
}

func StartBot(token string, locationService *MongoLocationsService, weather GetWeather) {
	botApi, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	//botApi.Debug = true

	bot := &TgBot{
		Api:      botApi,
		Location: locationService,
		Weather:  weather,
		chats:    make(map[int64]*Chat, 10),
	}

	bot.startBot()
}

type Result struct {
	MessageId int `json:"message_id"`
	Status string `json:"status"`
}

func toResponse(result Result, status int) events.APIGatewayProxyResponse {
	bytes, _ := json.Marshal(result)
	return events.APIGatewayProxyResponse { Body: string(bytes), StatusCode: status}
}
func (bot *TgBot) ListenForLambda() {
	lambda.Start(func (context context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		token, ok := request.PathParameters["token"]
		if !ok || token != os.Getenv("CALL_TOKEN") {
			return toResponse(Result{ Status: "invalid_token"}, 403), nil
		}

		update := tgbotapi.Update{}
		err := json.Unmarshal([]byte(request.Body), &update)
		if err != nil {
			return toResponse(Result{Status: "invalid_request"}, 400), nil
		}
		if update.Message != nil {
			chat, ok := bot.chats[update.Message.Chat.ID]
			if !ok {
				chat = NewChat(update.Message.Chat.ID, *bot)
				bot.chats[update.Message.Chat.ID] = chat
			}

			chat.ProcessMessage(update.Message)
			return toResponse(Result{Status: "ok", MessageId: update.Message.MessageID}, 200), nil
		} else {
			return toResponse(Result{Status: "not_a_message"}, 200), nil
		}
	})
}

func (bot *TgBot) startBot() {
	log.Printf("Authorized on account %s", bot.Api.Self.UserName)
	bot.ListenForLambda()
}

