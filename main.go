package main

import (
	"context"
	"encoding/json"
	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"tzbot/app"
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

	tzService, _ := initMongo()

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

func initMongo() (*app.MongoLocationsService, *mongo.Client) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	tzService := app.NewLocationsService(
		context.TODO(),
		client,
		func(c *mongo.Client) *mongo.Collection { return c.Database("tzbot").Collection("MONGO_TZ_COLLECTION") })

	return tzService, client
}
