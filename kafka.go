package main

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

func GetMyBrokerReq() (*MyBrokerReq, func()) {
	p, err := sarama.NewSyncProducer([]string{"grandma_kafka:9092"}, nil)
	if err != nil {
		logger.Fatalf("Error creating producer: %v", err)
	}
	consumer, err := sarama.NewConsumer([]string{"grandma_kafka:9092"}, nil)
	if err != nil {
		logger.Fatalf("Error creating consumer: %v", err)
	}
	c, err := consumer.ConsumePartition(ResponseTopic, 0, sarama.OffsetNewest)
	if err != nil {
		logger.Fatalf("Error partition: %v", err)
	}
	go func() {
		for msg := range c.Messages() {
			mu.Lock()
			ch, exists := responseChannels[string(msg.Key)]
			if exists {
				ch <- msg
			}
			mu.Unlock()
		}
		logger.Infof("Exiting goroutine...")
	}()

	return &MyBrokerReq{producer: p}, func() {
		p.Close()
		consumer.Close()
		c.Close()
	}
}

func GetMyBrokerResp(repo Repo) func() {
	consumer, err := sarama.NewConsumer([]string{"grandma_kafka:9092"}, nil)
	if err != nil {
		logger.Fatalf("Error creating consumer: %v", err)
	}
	p, err := sarama.NewSyncProducer([]string{"grandma_kafka:9092"}, nil)
	if err != nil {
		logger.Fatalf("Error creating producer: %v", err)
	}
	c, err := consumer.ConsumePartition(RequestsTopic, 0, sarama.OffsetNewest)
	if err != nil {
		logger.Fatalf("Error partition: %v", err)
	}
	go func() {
		for msg := range c.Messages() {
			var r Value
			err := json.Unmarshal(msg.Value, &r)
			if err != nil {
				logger.Errorf("Unmarshaling JSON: %v", err)
				continue
			}
			logger.Infof("Message: %+v\n", r)
			var respMsg string
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
			number, err := repo.GetNumber(ctx, r.Grandma)
			if err == nil {
				respMsg = strconv.Itoa(number)
			} else {
				r.Number = random.Int()
				err = repo.AddVal(ctx, r)
				if err != nil {
					respMsg = err.Error()
				} else {
					respMsg = strconv.Itoa(r.Number)
				}
			}
			cancel()
			bytes, err := json.Marshal(respMsg)
			if err != nil {
				logger.Errorf("Error marshal response message to Kafka: %v", err)
				continue
			}
			resp := &sarama.ProducerMessage{
				Topic: ResponseTopic,
				Key:   sarama.StringEncoder(msg.Key),
				Value: sarama.ByteEncoder(bytes),
			}
			_, _, err = p.SendMessage(resp)
			if err != nil {
				logger.Errorf("Error sending message to Kafka: %v", err)
			}
		}
	}()

	return func() {
		p.Close()
		consumer.Close()
		c.Close()
	}
}
