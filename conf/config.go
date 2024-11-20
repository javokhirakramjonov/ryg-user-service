package conf

import (
	"os"
)

type DBConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode    string
	TimeZone   string
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type Config struct {
	DB             DBConfig
	RabbitMQConfig RabbitMQConfig
	GRPCUrl        string
}

func LoadConfig() *Config {
	return &Config{
		DB: DBConfig{
			DBHost:     os.Getenv("DB_HOST"),
			DBPort:     os.Getenv("DB_PORT"),
			DBUser:     os.Getenv("DB_USER"),
			DBPassword: os.Getenv("DB_PASSWORD"),
			DBName:     os.Getenv("DB_NAME"),
			SSLMode:    os.Getenv("DB_SSL_MODE"),
			TimeZone:   os.Getenv("DB_TIMEZONE"),
		},
		GRPCUrl: os.Getenv("GRPC_URL"),
		RabbitMQConfig: RabbitMQConfig{
			Host:     os.Getenv("RABBITMQ_HOST"),
			Port:     os.Getenv("RABBITMQ_PORT"),
			User:     os.Getenv("RABBITMQ_USER"),
			Password: os.Getenv("RABBITMQ_PASSWORD"),
		},
	}
}
