package main

import (
	"os"
	"gotimezonebot/app"
	"gopkg.in/mgo.v2"
	"log"
	"github.com/go-redis/redis"
	"github.com/go-redis/cache"
	"encoding/json"
)

func main() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Panic("No TELEGRAM_TOKEN specified")
	}

	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	if weatherApiKey == "" {
		log.Panic("No WEATHER_API_KEY specified")
	}

	tzService, session := initMongo()
	defer session.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	pong, err := redisClient.Ping().Result()
	log.Println(pong, err)

	defer redisClient.Close()

	redisCache := &cache.Codec{
		Redis: redisClient,

		Marshal: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return json.Unmarshal(b, v)
		},
	}

	app.StartBot(telegramToken, tzService, app.WrapWeatherWithCache(app.CreateDarkSkyWeatherFetcher(weatherApiKey)), redisCache, false)
}

func initMongo() (*app.MongoLocationsService, *mgo.Session) {
	session, err := mgo.Dial(os.Getenv("MONGO_URL"))
	if err != nil {
		panic(err)
	}
	session.SetSafe(&mgo.Safe{})
	tzService := app.NewLocationsService(
		session,
		func(s *mgo.Session) *mgo.Collection { return s.DB("tzbot").C("MONGO_TZ_COLLECTION") })

	return tzService, session
}
