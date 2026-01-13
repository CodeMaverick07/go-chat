package email

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
}

func LoadConfig() *Config {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal("Invalid SMTP_PORT")
	}

	return &Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		FromName: os.Getenv("SMTP_FROM_NAME"),
	}
}