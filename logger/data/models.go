package data

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

//LogEntry is the model for the log entry in mongo
type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"` //string because we are working with mongo
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func New(mongo *mongo.Client) Models {
	client = mongo
	return Models{
		LogEntry: LogEntry{},
	}
}

func (l *LogEntry) Insert(entry LogEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(ctx, LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("error inserting into logs in mongodb: ", err)
		return err
	}
	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("error getting all logs from mongodb: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry
	for cursor.Next(ctx) {
		var logEntry LogEntry
		err := cursor.Decode(&logEntry)
		if err != nil {
			log.Println("error decoding log from mongodb: ", err)
			return nil, err
		}
		logs = append(logs, &logEntry)
	}

	return logs, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("logs").Collection("logs")

	documentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("error converting id in a document id format: ", err)
		return nil, err
	}

	var logEntry LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": documentID}).Decode(&logEntry)
	if err != nil {
		log.Println("error getting one log from mongodb: ", err)
		return nil, err
	}
	return &logEntry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("logs").Collection("logs")

	if err := collection.Drop(ctx); err != nil {
		log.Println("error dropping logs collection from mongodb: ", err)
		return err
	}
	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("logs").Collection("logs")

	documentID, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		log.Println("error converting id in a document id format: ", err)
		return nil, err
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": documentID},
		bson.D{{"$set", bson.D{
			{"name", l.Name},
			{"data", l.Data},
			{"updated_at", time.Now()},
		}},
		},
	)
	if err != nil {
		log.Println("error updating log in mongodb: ", err)
		return nil, err
	}

	return result, nil

}
