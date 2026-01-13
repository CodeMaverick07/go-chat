package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DB_HOST string
	DB_PORT string
	DB_USER string
	DB_PASSWORD string
	DB_NAME string

	JWTSecret     string
	JWTExpiryMins int

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func Load() *Config {
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal("Invalid SMTP_PORT")
	}

	jwtExp, _ := strconv.Atoi(os.Getenv("JWT_EXP_MINUTES"))

	return &Config{
		DB_HOST: os.Getenv("DB_HOST"),
		DB_PORT: os.Getenv("DB_PORT"),
		DB_USER: os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_NAME: os.Getenv("DB_NAME"),

		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiryMins: jwtExp,

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: smtpPort,
		SMTPUser: os.Getenv("SMTP_USERNAME"),
		SMTPPass: os.Getenv("SMTP_PASSWORD"),
	}
}