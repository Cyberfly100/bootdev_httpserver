package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func readDBURL() string {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	return dbURL
}

func readJWTSecret() string {
	godotenv.Load()
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	return JWTSecret
}

func readPlatform() string {
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM environment variable is not set")
	}
	return platform
}

func readPolkaAPIKey() string {
	polkaAPIKey := os.Getenv("POLKA_API_KEY")
	if polkaAPIKey == "" {
		log.Fatal("POLKA_API_KEY environment variable is not set")
	}
	return polkaAPIKey
}
