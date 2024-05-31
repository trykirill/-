package main

import (
	"context"
	"math/rand"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var responseChannels = make(map[string]chan *sarama.ConsumerMessage)
var mu sync.Mutex
var broker Broker
var random *rand.Rand

func main() {
	random = rand.New(rand.NewSource(999))
	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		FullTimestamp: true,
	})
	logger.SetReportCaller(true)
	ctx, c3 := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer c3()
	repo, client := NewRepo(ctx)
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Fatal(err)
		}
	}()
	b, c1 := GetMyBrokerReq()
	broker.b = b
	defer c1()
	c2 := GetMyBrokerResp(*repo)
	defer c2()
	server := Server()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Listen error: %v", err)
		}
	}()
	logger.Info("Start listen :9000")

	<-ctx.Done()
	shtCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	server.Shutdown(shtCtx)

}
