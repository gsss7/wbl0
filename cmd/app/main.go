package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"wbl0/internal/cache"
	"wbl0/internal/config"
	"wbl0/internal/db"
	httpx "wbl0/internal/http"
	"wbl0/internal/kafka"
	"wbl0/internal/models"
	"wbl0/internal/repo"
	validatorx "wbl0/internal/validator"
)

func runHTTP(addr string, h http.Handler) error {
	log.Printf("http server listening on %s", addr)
	return http.ListenAndServe(addr, h)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func main() {
	cfgPath := envOr("CONFIG", "./config.local.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	validatorx.Init()

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DB, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.MaxConns)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	repository := repo.New(pool)

	c := cache.New[models.Order](cfg.Cache.TTL, cfg.Cache.MaxItems)
	defer c.Stop()

	if cfg.Cache.PreloadLastN > 0 {
		if orders, err := repository.LoadRecentOrders(ctx, cfg.Cache.PreloadLastN); err != nil {
			for _, o := range orders {
				c.Set(o.OrderUID, *o)
			}
			log.Printf("cache preloaded: %d entries", len(orders))
		} else {
			log.Printf("cache preload failed: %v", err)
		}
	}

	r := httpx.NewRouter()
	api := &httpx.API{Repo: repository, Cache: c}
	api.Register(r)
	httpx.MountStatic(r, "./assets")

	srvErr := make(chan error, 1)
	go func() { srvErr <- runHTTP(cfg.HTTP.Addr, r) }()

	commitEvery := time.Duration(cfg.Kafka.CommitInterval) * time.Millisecond
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, commitEvery, repository,
		func(o *models.Order) { c.Invalidate(o.OrderUID); c.Set(o.OrderUID, *o) })

	consErr := make(chan error, 1)
	go func() { consErr <- consumer.Run(ctx) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-sig:
		log.Printf("signal: %v", s)
	case err := <-srvErr:
		log.Printf("http error: %v", err)
	case err := <-consErr:
		log.Printf("kafka error: %v", err)
	}

	_ = consumer.Close()
}
