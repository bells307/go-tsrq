package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	GrpcListen string      `mapstructure:"GRPC_LISTEN"`
	RedisAddr  string      `mapstructure:"REDIS_ADDR"`
	Queue      QueueConfig `mapstructure:"QUEUE"`
}

type QueueConfig struct {
	Name        string        `mapstructure:"NAME" validate:"required"`
	DeqPeriod   time.Duration `mapstructure:"DEQ_PERIOD" validate:"required"`
	Ttl         time.Duration `mapstructure:"TTL" validate:"required"`
	CleanPeriod time.Duration `mapstructure:"CLEAN_PERIOD"`
}

func LoadConfig(path string) (cfg Config, err error) {
	viper.SetConfigFile(path)

	viper.SetDefault("grpc_listen", ":9999")
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("queue.clean_period", 10*time.Second)

	err = viper.ReadInConfig()
	if err = viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("error while reading config file: %v", err)
	}

	if err = viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("error while unmarshalling config file: %v", err)
	}

	validate := validator.New()
	if err = validate.Struct(&cfg); err != nil {
		return Config{}, fmt.Errorf("error while validating config file: %v", err)
	}

	return
}
