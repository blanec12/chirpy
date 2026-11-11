package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/blanec12/chirpy/internal/database"
	"github.com/joho/godotenv"
	"os"
	"sync/atomic"
)

const PORT string = ":8080"

type Config struct {
	platform       string
	db             *database.Queries
	port           string
	fileServerHits atomic.Int32
	jwtSecret      string
	polkaKey       string
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("Error loading .env: %w", err)
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	if dbURL == "" {
		return nil, errors.New("DB_URL environment variable is required")
	}

	if platform == "" {
		return nil, errors.New("PLATFORM environment variable is required")
	}

	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	if polkaKey == "" {
		return nil, errors.New("POLKA_KEY environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %w", err)
	}

	config := &Config{
		platform:  platform,
		db:        database.New(db),
		port:      PORT,
		jwtSecret: jwtSecret,
		polkaKey:  polkaKey,
	}

	return config, nil
}
