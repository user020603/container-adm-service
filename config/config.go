package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	GrpcPort      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisAddr     string
	RedisPassword string
	EsAddr        string
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroupID  string
	LogLevel      string
	LogFile       string
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		configInstance = &Config{
			ServerPort:    getEnv("SERVER_PORT", "8001"),
			GrpcPort:      getEnv("GRPC_PORT", "50051"),
			DBHost:        getEnv("DB_HOST", "localhost"),
			DBPort:        getEnv("DB_PORT", "5432"),
			DBUser:        getEnv("DB_USER", "postgres"),
			DBPassword:    getEnv("DB_PASSWORD", "password"),
			DBName:        getEnv("DB_NAME", "admdb"),
			RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			EsAddr:        getEnv("ES_ADDR", "http://localhost:9200"),
			KafkaBrokers:  []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			KafkaTopic:    getEnv("KAFKA_TOPIC", "container_topic"),
			KafkaGroupID:  getEnv("KAFKA_GROUP_ID", "container_group_id"),
			LogLevel:      getEnv("LOG_LEVEL", "info"),
			LogFile:       getEnv("LOG_FILE", "../logs/container.log"),
		}
	})

	return configInstance
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
