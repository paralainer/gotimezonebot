package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Location struct {
	Alias               string `bson:"alias"`
	Coordinates string `bson:"coordinates"`
}

type tz struct {
	ChatID    int64 `bson:"chatId"`
	Locations []Location `bson:"locations"`
}

type MongoTimezonesService struct {
	Collection *mgo.Collection
}

func NewTzService(mongoCollection *mgo.Collection) *MongoTimezonesService {
	return &MongoTimezonesService{
		Collection: mongoCollection,
	}
}

func (botTz MongoTimezonesService) GetChatTimezones(chatID int64) []Location {
	tz := &tz{}
	botTz.Collection.Find(bson.M{"chatId": chatID}).One(&tz)
	return tz.Locations
}

func (botTz MongoTimezonesService) AddTimezone(chatID int64, location *Location) {
	count, _ := botTz.Collection.Find(bson.M{"chatId": chatID}).Count()
	if count == 0 {
		tz := &tz{
			ChatID:    chatID,
			Locations: []Location{*location},
		}

		botTz.Collection.Insert(tz)
	} else {
		//botTz.Collection.Update(bson.M{"chatId": chatID}, )
	}
}

func (botTz MongoTimezonesService) RemoveTimezone(chatID int64, alias string) {

}
