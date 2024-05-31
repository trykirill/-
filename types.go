package main

import (
	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repo struct {
	db *mongo.Collection
}
type Value struct {
	Grandma string `bson:"username"`
	Number  int    `bson:"number"`
}

type Message struct {
	ID    string `json:"id"`
	Value Value  `json:"value"`
}

type MyBrokerReq struct {
	producer sarama.SyncProducer
}

type Broker struct {
	b *MyBrokerReq
}

const (
	RequestsTopic = "1"
	ResponseTopic = "2"
)