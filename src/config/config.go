package config

import "flag"

type Config struct {
	Port      string
	JWTSecret string
	DSN       string
}

var cfg *Config

func Get() *Config {
	if cfg != nil {
		return cfg
	}
	port      := flag.String("port", "4000", "Port in which the Flagpole API will serve")
	jwtSecret := flag.String("jwt-secret", "change-me", "Secret key used to sign JWT tokens")
	dsn       := flag.String("dsn", "", "PostgreSQL DSN (e.g. host=localhost user=postgres password=postgres dbname=flagpole port=5432 sslmode=disable)")
	flag.Parse()
	cfg = &Config{Port: *port, JWTSecret: *jwtSecret, DSN: *dsn}
	return cfg
}
