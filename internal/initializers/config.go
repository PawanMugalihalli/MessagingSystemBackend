package initializers

import (
	"log"
	"github.com/joho/godotenv"
)

var JWT_SECRET string

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
