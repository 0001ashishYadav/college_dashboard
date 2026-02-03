package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	TokenSymmetricKey  string
	Port               int16
	TokenDuration      time.Duration
	ProfilesFolder     string
	MqttBroker         string
	MqttClientID       string
	MqttUser           string
	MqttPassword       string
	AttendanceDeviceID string
}

func LoadConfig(path string) (Config, error) {
	// Load .env file only in local development
	_ = godotenv.Load(path + "/.env") // Ignore error if file doesn't exist

	tokenDurationStr := os.Getenv("TOKEN_DURATION")
	tokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil || tokenDurationStr == "" {
		tokenDuration = time.Hour // default value if not set or invalid
	}

	portStr := os.Getenv("PORT")
	portInt, err := strconv.Atoi(portStr)
	if err != nil || portStr == "" {
		portInt = 8080 // default port if not set or invalid
	}
	port := int16(portInt)

	config := Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		TokenSymmetricKey:  os.Getenv("TOKEN_SYMMETRIC_KEY"),
		Port:               port,
		TokenDuration:      tokenDuration,
		ProfilesFolder:     os.Getenv("PROFILES_FOLDER"),
		MqttBroker:         os.Getenv("MQTT_BROKER"),
		MqttClientID:       os.Getenv("MQTT_CLIENT_ID"),
		MqttUser:           os.Getenv("MQTT_USER"),
		MqttPassword:       os.Getenv("MQTT_PASSWORD"),
		AttendanceDeviceID: os.Getenv("ATTENDANCE_DEVICE_ID"),
	}

	// Optionally, check for required variables
	if config.DatabaseURL == "" || config.TokenSymmetricKey == "" || config.MqttBroker == "" || config.MqttClientID == "" || config.MqttUser == "" || config.MqttPassword == "" || config.AttendanceDeviceID == "" {
		return config, ErrMissingEnv
	}

	return config, nil
}

var ErrMissingEnv = &ConfigError{"One or more required environment variables are missing"}

type ConfigError struct {
	s string
}

func (e *ConfigError) Error() string {
	return e.s
}
