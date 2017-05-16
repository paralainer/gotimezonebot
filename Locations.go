package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type Location struct {
	Alias       string `bson:"alias"`
	Coordinates string `bson:"coordinates"`
}

type tz struct {
	ChatID    int64 `bson:"chatId"`
	Locations []Location `bson:"locations"`
}

type MongoTimezonesService struct {
	collection *mgo.Collection
}

func NewTzService(mongoCollection *mgo.Collection) *MongoTimezonesService {
	return &MongoTimezonesService{
		collection: mongoCollection,
	}
}

func (botTz MongoTimezonesService) GetChatLocations(chatID int64) []Location {
	tz := &tz{}
	botTz.collection.Find(bson.M{"chatId": chatID}).One(&tz)
	return tz.Locations
}

func (botTz MongoTimezonesService) AddLocation(chatID int64, location Location) {
	count, _ := botTz.collection.Find(bson.M{"chatId": chatID}).Count()
	if count == 0 {
		tz := &tz{
			ChatID:    chatID,
			Locations: []Location{location},
		}

		botTz.collection.Insert(tz)
	} else {
		//botTz.Collection.Update(bson.M{"chatId": chatID}, )
	}
}

func (botTz MongoTimezonesService) RemoveTimezone(chatID int64, alias string) bool {
	findQuery := bson.M{"chatId": chatID, "locations": bson.M{"$elemMatch": bson.M{"alias": alias}}}
	updateQuery := bson.M{"$pull": bson.M{"locations": bson.M{"alias": alias}}}
	err := botTz.collection.Update(findQuery, updateQuery)
	if err == mgo.ErrNotFound {
		return false
	}

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
