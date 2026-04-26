package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	JWTSecret   string
	DSN         string
	Env         string
	AllowOrigin string
}

var cfg Config

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading config from environment")
	}

	cfg = Config{
		Port:        getEnv("PORT", "4000"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me"),
		DSN:         getEnv("DSN", ""),
		Env:         getEnv("ENV", ""),
		AllowOrigin: getEnv("ALLOW_ORIGIN", "http://localhost:5173"),
	}

	port      := flag.String("port", "", "Port in which the flagpole API will serve")
	jwtSecret := flag.String("jwt-secret", "", "Secret key used to sign JWT tokens")
	dsn       := flag.String("dsn", "", "PostgreSQL DSN (e.g. host=localhost user=postgres password=postgres dbname=flagpole port=5432 sslmode=disable)")
	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			cfg.Port = *port
		case "jwt-secret":
			cfg.JWTSecret = *jwtSecret
		case "dsn":
			cfg.DSN = *dsn
		}
	})
}

func Get() Config {
	return cfg
}
