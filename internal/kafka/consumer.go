package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"wbl0/internal/models"
	"wbl0/internal/repo"
	validatorx "wbl0/internal/validator"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r       *kafka.Reader
	repo    *repo.Repository
	onOrder func(*models.Order)
}

func NewConsumer(brokers []string, topic, group string, commitInterval time.Duration, repo *repo.Repository, onOrder func(order *models.Order)) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  group,
		Topic:    topic,
		MinBytes: 1, MaxBytes: 10e6,
		CommitInterval: commitInterval,
	})
	return &Consumer{r: r, repo: repo, onOrder: onOrder}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			return err
		}

		var ord models.Order
		if err := json.Unmarshal(m.Value, &ord); err != nil {
			log.Printf("[kafka] invalid JSON, offset=%d: %v", m.Offset, err)
			_ = c.r.CommitMessages(ctx, m)
			continue
		}

		if err := validatorx.V().Struct(ord); err != nil {
			log.Printf("[kafka] validation failed, offset=%d: %v", m.Offset, err)
			_ = c.r.CommitMessages(ctx, m)
			continue
		}

		saveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err = c.repo.InsertOrUpdateOrder(saveCtx, &ord)
		cancel()

		if err != nil {
			log.Printf("[kafka] db error, will retry offset=%d: %v", m.Offset, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if c.onOrder != nil {
			c.onOrder(&ord)
		}

		if err := c.r.CommitMessages(ctx, m); err != nil {
			log.Printf("[kafka] commit failed: %v", err)
		}
	}
}

func (c *Consumer) Close() error { return c.r.Close() }
