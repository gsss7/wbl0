package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTP struct {
	Addr string `yaml:"addr" env:"HTTP_ADDR" env-default:":8081"`
}

type Postgres struct {
	Host     string `yaml:"host" env:"PG_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"PG_PORT" env-default:"5432"`
	DB       string `yaml:"db" env:"PG_DB" env-default:"orders"`
	User     string `yaml:"user" env:"PG_USER" env-default:"user"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-default:"password"`
	MaxConns int    `yaml:"max_conns" env:"PG_MAX_CONNS" env-default:"10"`
}

type Kafka struct {
	Brokers        []string `yaml:"brokers" env:"KAFKA_BROKERS" env-separator:","`
	Topic          string   `yaml:"topic" env:"KAFKA_TOPIC" env-default:"orders"`
	GroupID        string   `yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"orders-consumer"`
	CommitInterval int      `yaml:"commit_interval_ms" env:"KAFKA_COMMIT_INTERVAL_MS" env-default:"1000"`
}

type Cache struct {
	TTL          time.Duration `yaml:"ttl" env:"CACHE_TTL" env-default:"10m"`
	MaxItems     int           `yaml:"max_items" env:"CACHE_MAX_ITEMS" env-default:"10000"`
	PreloadLastN int           `yaml:"preload_last_n" env:"CACHE_PRELOAD_LAST_N" env-default:"100"`
}

type Config struct {
	App struct {
		Name string `yaml:"name" env:"APP_NAME" env-default:"orders-service"`
		Env  string `yaml:"env" env:"APP_ENV" env-default:"local"`
	} `yaml:"app"`

	HTTP     HTTP     `yaml:"http"`
	Postgres Postgres `yaml:"postgres"`
	Kafka    Kafka    `yaml:"kafka"`
	Cache    Cache    `yaml:"cache"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
