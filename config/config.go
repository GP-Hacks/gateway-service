package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env               string
	LocalAddress      string
	Address           string
	ChatAddress       string
	PlacesAddress     string
	CharityAddress    string
	VotesAddress      string
	AuthAddress       string
	UsersAddress      string
	Timeout           time.Duration
	IdleTimeout       time.Duration
	MongoDBName       string
	MongoDBCollection string
	MongoDBPath       string
	KafkaBrokers      []string
	RequestTopic      string
	ResponseTopic     string
	ResponseTimeout   time.Duration
}

func MustLoad() *Config {
	return &Config{
		Env:               getEnv("ENV", "local"),
		Address:           getEnv("SERVICE_ADDRESS", ":8080"),
		LocalAddress:      getEnv("LOCAL_ADDRESS", "0.0.0.0:8080"),
		ChatAddress:       getEnv("CHAT_SERVICE_ADDRESS", ""),
		PlacesAddress:     getEnv("PLACES_SERVICE_ADDRESS", ""),
		CharityAddress:    getEnv("CHARITY_SERVICE_ADDRESS", ""),
		VotesAddress:      getEnv("VOTES_SERVICE_ADDRESS", ""),
		AuthAddress:       getEnv("AUTH_SERVICE_ADDRESS", ""),
		UsersAddress:      getEnv("USERS_SERVICE_ADDRESS", ""),
		Timeout:           getDurationEnv("TIMEOUT", time.Second*15),
		IdleTimeout:       getDurationEnv("IDLE_TIMEOUT", time.Second*60),
		MongoDBName:       getEnv("MONGODB_NAME", ""),
		MongoDBCollection: getEnv("MONGODB_COLLECTION", ""),
		MongoDBPath:       getEnv("MONGODB_PATH", ""),
		KafkaBrokers:      getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
		RequestTopic:      getEnv("KAFKA_REQUEST_TOPIC", "request_topic"),
		ResponseTopic:     getEnv("KAFKA_RESPONSE_TOPIC", "response_topic"),
		ResponseTimeout:   getDurationEnv("KAFKA_RESPONSE_TIMEOUT", time.Second*30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(strings.TrimSpace(value), ",")
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
