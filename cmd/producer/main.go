package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	broker := env("BROKER", "localhost:9094")
	topic := env("TOPIC", "orders")
	file := env("FILE", "./model.json")

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var orders []map[string]interface{}
	if err := json.Unmarshal(data, &orders); err != nil {
		log.Fatal("failed to unmarshal JSON:", err)
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.LeastBytes{},
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, ord := range orders {
		b, _ := json.Marshal(ord) // каждое сообщение — один заказ
		if err := w.WriteMessages(ctx, kafka.Message{Value: b}); err != nil {
			log.Fatal(err)
		}
		log.Printf("sent order %s (%d bytes)", ord["order_uid"], len(b))
	}
}
