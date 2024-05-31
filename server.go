package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetNumber(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("username") == "" {
		logger.Warn("Bad Request, missed 'username' key")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id := uuid.New().String()
	value := Value{Grandma: r.URL.Query().Get("username")}
	b, err := json.Marshal(value)
	if err != nil {
		logger.Errorf("Marshal JSON err: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	msg := &sarama.ProducerMessage{
		Topic: RequestsTopic,
		Key:   sarama.StringEncoder(id),
		Value: sarama.ByteEncoder(b),
	}

	_, _, err = broker.b.producer.SendMessage(msg)
	if err != nil {
		logger.Errorf("Kafka send err: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	c := make(chan *sarama.ConsumerMessage)
	mu.Lock()
	responseChannels[id] = c
	mu.Unlock()
	defer f(id)
	select {
	case responseMsg := <- c:
		var r string
		err := json.Unmarshal(responseMsg.Value, &r)
		if err != nil {
			logger.Error("Error of unmarshaling json")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logger.Infof("Service Response: %v", r)
		w.WriteHeader(http.StatusOK)
		json, err := json.Marshal(r)
		if err != nil {
			logger.Errorf("Error Marshaling JSON: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(json)
		if err != nil {
			logger.Errorf("Error writing JSON: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	case <-time.After(2 * time.Second):
		logger.Warnf("Request processing takes too long")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func Server() http.Server {
	return http.Server{
		Addr:         ":9000",
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		Handler:      api(),
	}
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info( "Url req - ", r.URL.Path, " method - ", r.Method)
		logger.Infof("time: %v", time.Since(start))
	})
}

func f(id string) {
	mu.Lock()
	close(responseChannels[id])
	delete(responseChannels, id)
	mu.Unlock()
}

func api() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/api/get-number", GetNumber).Methods("POST")
	r.Use(func(hdl http.Handler) http.Handler {
		return Log(hdl)
	})

	return r
}

