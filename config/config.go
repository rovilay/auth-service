package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type AppConfig struct {
	ServerPort   uint16
	DATABASE_URL string
	JwtSecret    string
}

var Config = AppConfig{}

func LoadConfig(log *zerolog.Logger) AppConfig {
	Config.ServerPort = 3000

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			Config.ServerPort = uint16(port)
		}
	}

	if url, exists := os.LookupEnv("DATABASE_URL"); exists {
		Config.DATABASE_URL = url
	}

	if secret, exists := os.LookupEnv("JWT_SECRET"); exists {
		Config.JwtSecret = secret
	} else {
		log.Fatal().Err(errors.New("JWT_SECRET is required")).Msg("failed to load config")
	}

	return Config
}
