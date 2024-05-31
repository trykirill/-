package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

func NewRepo(ctx context.Context) (*Repo, *mongo.Client) {
	uri := "mongodb://grandma_mongo:27017/messenger?directConnection=true"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		logger.Fatalf("Error: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatalf("Error: %s", err)
	}

	collection := client.Database("grandma").Collection("numbers")
	return &Repo{db: collection}, client
}

func (repo *Repo) GetNumber(ctx context.Context, name string) (int, error) {
	res := repo.db.FindOne(ctx, bson.M{"username": name})
	if res.Err() == mongo.ErrNoDocuments {
		return 0, res.Err()
	} else if res.Err() != nil {
		return 0, fmt.Errorf("get user mongo error: %w", res.Err())
	}
	var value Value
	err := res.Decode(&value)
	if err != nil {
		return 0, fmt.Errorf("get user mongo error (decode): %w", err)
	}
	return value.Number, nil
}

func (repo *Repo) AddVal(ctx context.Context, val Value) error {
	_, err := repo.db.InsertOne(ctx, val)
	if err != nil {
		return fmt.Errorf("insert mongo error: %w", err)
	}
	return nil
}
