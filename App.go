package main

import (
	"log"
	"os"
	"gopkg.in/mgo.v2"
)

func main() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Panic("No TELEGRAM_TOKEN specified")
	}

	tzService, session := initMongo()
	defer session.Close()

	StartBot(telegramToken, tzService, GetDarkSkyWeather);
}

func initMongo() (*MongoTimezonesService, *mgo.Session) {
	session, err := mgo.Dial(os.Getenv("MONGO_URL"))
	if err != nil {
		panic(err)
	}
	c := session.DB("tzbot").C("MONGO_TZ_COLLECTION")
	tzService := NewTzService(c)
	return tzService, session
}
