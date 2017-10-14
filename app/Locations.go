package app

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

type MongoLocationsService struct {
	session *mgo.Session
	getCollection GetCollection
}

type GetCollection func(session *mgo.Session) *mgo.Collection

func NewLocationsService(mongoSession *mgo.Session, getCollection GetCollection) *MongoLocationsService {
	return &MongoLocationsService{
		session: mongoSession,
		getCollection: getCollection,
	}
}

func (botTz MongoLocationsService) collection() (*mgo.Collection, *mgo.Session) {
    session := botTz.session.Copy()
	return botTz.getCollection(session), session
}

func (botTz MongoLocationsService) GetChatLocations(chatID int64) []Location {
	tz := &tz{}
	collection, session := botTz.collection()
	defer session.Close()
	err := collection.Find(bson.M{"chatId": chatID}).One(&tz)
	if err != nil {
		log.Println(err)
	}
	return tz.Locations
}

func (botTz MongoLocationsService) AddLocation(chatID int64, location Location) {
	collection, session := botTz.collection()
	defer session.Close()
	count, _ := collection.Find(bson.M{"chatId": chatID}).Count()
	if count == 0 {
		tz := &tz{
			ChatID:    chatID,
			Locations: []Location{location},
		}

		collection.Insert(tz)
	} else {
		collection.Update(bson.M{"chatId": chatID}, bson.M{"$addToSet": bson.M{"locations": location}} )
	}
}

func (botTz MongoLocationsService) RemoveTimezone(chatID int64, alias string) bool {
	findQuery := bson.M{"chatId": chatID, "locations": bson.M{"$elemMatch": bson.M{"alias": alias}}}
	updateQuery := bson.M{"$pull": bson.M{"locations": bson.M{"alias": alias}}}
	collection, session := botTz.collection()
	defer session.Close()
	err := collection.Update(findQuery, updateQuery)
	if err == mgo.ErrNotFound {
		return false
	}

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
