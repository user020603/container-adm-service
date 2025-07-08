package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetConfig() {
	once = sync.Once{}
	configInstance = nil
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetConfig()
	os.Clearenv()

	cfg := LoadConfig()

	assert.Equal(t, "8001", cfg.ServerPort)
	assert.Equal(t, "50051", cfg.GrpcPort)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "password", cfg.DBPassword)
	assert.Equal(t, "serverdb", cfg.DBName)
	assert.Equal(t, "localhost:6379", cfg.RedisAddr)
	assert.Equal(t, "", cfg.RedisPassword)
	assert.Equal(t, "http://localhost:9200", cfg.EsAddr)
	assert.Equal(t, []string{"localhost:9092"}, cfg.KafkaBrokers)
	assert.Equal(t, "container_status", cfg.KafkaTopic)
	assert.Equal(t, "container_group_id", cfg.KafkaGroupID)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "../../logs/container-adm.log", cfg.LogFile)
	assert.Equal(t, "supersecretkey", cfg.JWTSecret)
	assert.Equal(t, 3600, cfg.JWTExpiresIn)
	assert.Equal(t, 604800, cfg.RefreshTokenTTL)
}

func TestLoadConfig_FromEnvironment(t *testing.T) {
	resetConfig()

	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("GRPC_PORT", "6000")
	os.Setenv("DB_HOST", "db.internal")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "securepass")
	os.Setenv("DB_NAME", "productiondb")
	os.Setenv("REDIS_ADDR", "redis.internal:6380")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("ELASTICSEARCH_URL", "http://es.internal:9200")
	os.Setenv("KAFKA_BROKERS", "broker1:9092")
	os.Setenv("KAFKA_TOPIC", "logs_topic")
	os.Setenv("KAFKA_GROUP_ID", "log_group")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FILE", "/var/log/container.log")
	os.Setenv("JWT_SECRET", "envsecret")
	os.Setenv("JWT_EXPIRES_IN", "7200")
	os.Setenv("REFRESH_TOKEN_TTL", "1209600")

	cfg := LoadConfig()

	assert.Equal(t, "9000", cfg.ServerPort)
	assert.Equal(t, "6000", cfg.GrpcPort)
	assert.Equal(t, "db.internal", cfg.DBHost)
	assert.Equal(t, "3306", cfg.DBPort)
	assert.Equal(t, "admin", cfg.DBUser)
	assert.Equal(t, "securepass", cfg.DBPassword)
	assert.Equal(t, "productiondb", cfg.DBName)
	assert.Equal(t, "redis.internal:6380", cfg.RedisAddr)
	assert.Equal(t, "redispass", cfg.RedisPassword)
	assert.Equal(t, "http://es.internal:9200", cfg.EsAddr)
	assert.Equal(t, []string{"broker1:9092"}, cfg.KafkaBrokers)
	assert.Equal(t, "logs_topic", cfg.KafkaTopic)
	assert.Equal(t, "log_group", cfg.KafkaGroupID)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "/var/log/container.log", cfg.LogFile)
	assert.Equal(t, "envsecret", cfg.JWTSecret)
	assert.Equal(t, 7200, cfg.JWTExpiresIn)
	assert.Equal(t, 1209600, cfg.RefreshTokenTTL)
}

func TestLoadConfig_InvalidIntegers(t *testing.T) {
	resetConfig()

	os.Setenv("JWT_EXPIRES_IN", "not_a_number")
	os.Setenv("REFRESH_TOKEN_TTL", "invalid")

	cfg := LoadConfig()

	assert.Equal(t, 3600, cfg.JWTExpiresIn)
	assert.Equal(t, 604800, cfg.RefreshTokenTTL)
}

func TestLoadConfig_SingletonBehavior(t *testing.T) {
	resetConfig()

	os.Setenv("SERVER_PORT", "9001")
	cfg1 := LoadConfig()
	cfg2 := LoadConfig()

	assert.Same(t, cfg1, cfg2, "LoadConfig should return the same instance on subsequent calls")
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_ENV_KEY", "some_value")
	assert.Equal(t, "some_value", getEnv("TEST_ENV_KEY", "default"))

	os.Unsetenv("TEST_ENV_KEY")
	assert.Equal(t, "default", getEnv("TEST_ENV_KEY", "default"))
}
