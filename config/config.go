package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	GrpcPort        string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	RedisAddr       string
	RedisPassword   string
	EsAddr          string
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	LogLevel        string
	LogFile         string
	JWTSecret       string
	JWTExpiresIn    int
	RefreshTokenTTL int
}

var (
	configInstance *Config
	once           sync.Once
)

var LoadConfig = func() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		jwtExpiresIn, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN", "3600"))
		if err != nil {
			jwtExpiresIn = 3600
		}
		refreshTokenTTL, err := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL", "604800"))
		if err != nil {
			refreshTokenTTL = 604800
		}

		configInstance = &Config{
			ServerPort:      getEnv("SERVER_PORT", "8001"),
			GrpcPort:        getEnv("GRPC_PORT", "50051"),
			DBHost:          getEnv("DB_HOST", "localhost"),
			DBPort:          getEnv("DB_PORT", "5432"),
			DBUser:          getEnv("DB_USER", "postgres"),
			DBPassword:      getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "admdb"),
			RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPassword:   getEnv("REDIS_PASSWORD", ""),
			EsAddr:          getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
			KafkaBrokers:    []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			KafkaTopic:      getEnv("KAFKA_TOPIC", "container_topic"),
			KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "container_group_id"),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			LogFile:         getEnv("LOG_FILE", "../../logs/container-adm.log"),
			JWTSecret:       getEnv("JWT_SECRET", "supersecretkey"),
			JWTExpiresIn:    jwtExpiresIn,
			RefreshTokenTTL: refreshTokenTTL,
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
