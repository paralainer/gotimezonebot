package app

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Location struct {
	Alias       string `bson:"alias"`
	Coordinates string `bson:"coordinates"`
}

type tz struct {
	ChatID    int64      `bson:"chatId"`
	Locations []Location `bson:"locations"`
}

type MongoLocationsService struct {
	ctx           context.Context
	client        *mongo.Client
	getCollection GetCollection
}

type GetCollection func(session *mongo.Client) *mongo.Collection

func NewLocationsService(ctx context.Context, mongoClient *mongo.Client, getCollection GetCollection) *MongoLocationsService {
	return &MongoLocationsService{
		ctx:           ctx,
		client:        mongoClient,
		getCollection: getCollection,
	}
}

func (botTz MongoLocationsService) collection() *mongo.Collection {
	return botTz.getCollection(botTz.client)
}

func (botTz MongoLocationsService) GetChatLocations(chatID int64) []Location {
	tz := &tz{}
	collection := botTz.collection()
	err := collection.FindOne(botTz.ctx, bson.M{"chatId": chatID}).Decode(&tz)
	if err != nil {
		log.Println(err)
	}
	return tz.Locations
}

func (botTz MongoLocationsService) AddLocation(chatID int64, location Location) {
	collection := botTz.collection()
	count, _ := collection.CountDocuments(botTz.ctx, bson.M{"chatId": chatID})
	if count == 0 {
		tz := &tz{
			ChatID:    chatID,
			Locations: []Location{location},
		}

		collection.InsertOne(botTz.ctx, tz)
	} else {
		collection.UpdateOne(botTz.ctx, bson.M{"chatId": chatID}, bson.M{"$addToSet": bson.M{"locations": location}})
	}
}

func (botTz MongoLocationsService) RemoveTimezone(chatID int64, alias string) bool {
	findQuery := bson.M{"chatId": chatID, "locations": bson.M{"$elemMatch": bson.M{"alias": alias}}}
	updateQuery := bson.M{"$pull": bson.M{"locations": bson.M{"alias": alias}}}
	collection := botTz.collection()
	result, err := collection.UpdateOne(botTz.ctx, findQuery, updateQuery)
	if result != nil && result.MatchedCount == 0 {
		return false
	}

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
