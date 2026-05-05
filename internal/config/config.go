package config

import "os"

type Config struct {
	TelegramToken string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	DiscordToken string `envconfig:"DISCORD_TOKEN" required:"true"`
}

func NewConfig() (*Config, error) {
	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
	}, nil
}