package app

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/aws/aws-lambda-go/lambda"
	"context"
	"os"
	"github.com/aws/aws-lambda-go/events"
	"github.com/go-redis/cache"
	"encoding/json"
	"time"
	"net/http"
	"strconv"
)

type TgBot struct {
	Location *MongoLocationsService
	Api      *tgbotapi.BotAPI
	Weather  GetWeather
	redisCache *cache.Codec
	local bool
}

func StartBot(token string, locationService *MongoLocationsService, weather GetWeather, redisCache *cache.Codec, local bool) {
	botApi, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	//botApi.Debug = true

	bot := &TgBot{
		Api:      botApi,
		Location: locationService,
		Weather:  weather,
		redisCache: redisCache,
		local: local,
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
		return toResponse(bot.handleUpdate(&update), 200), nil
	})
}

func (bot *TgBot) handleUpdate(update *tgbotapi.Update) (Result) {
	if update.Message != nil {
		chatState, ok := bot.GetChat(update.Message.Chat.ID)
		var chat *Chat
		if !ok {
			chat = NewChat(update.Message.Chat.ID, bot)
		} else {
			chat = RestoreChat(chatState, bot)
		}

		chat.ProcessMessage(update.Message)

		bot.SetChat(update.Message.Chat.ID, chat.State)
		return Result{Status: "ok", MessageId: update.Message.MessageID}
	} else {
		return Result{Status: "not_a_message"}
	}
}

func (bot *TgBot) GetChat(chatId int64) (*ChatState, bool){
	var key = strconv.FormatInt(chatId, 10)
	if bot.redisCache.Exists(key) {
		chat := ChatState{}
		err := bot.redisCache.Get(key, &chat)
		if err != nil {
			log.Println(err)
			return nil, false
		}
		return &chat, true
	} else {
		return nil, false
	}
}

func (bot *TgBot) SetChat(chatId int64, chat *ChatState) {
	bot.redisCache.Set(&cache.Item{
		Key: strconv.FormatInt(chatId, 10),
		Object: chat,
		Expiration: time.Hour,
	})
}



func (bot *TgBot) startBot() {
	log.Printf("Authorized on account %s", bot.Api.Self.UserName)
	if bot.local {
		http.HandleFunc("/Bot-callback", func(writer http.ResponseWriter, request *http.Request) {
			update := tgbotapi.Update{}
			err := json.NewDecoder(request.Body).Decode(&update)
			if err != nil {
				log.Println(err)
				writer.Write([]byte(err.Error()))
				writer.WriteHeader(500)
			} else {
				result := bot.handleUpdate(&update)
				bytes, _ := json.Marshal(&result)
				writer.Write(bytes)
				writer.WriteHeader(200)
			}
		})
		http.ListenAndServe("0.0.0.0:3456", nil)
	} else {
		bot.ListenForLambda()
	}
}

