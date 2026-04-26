# flagpole-server

REST API for the flagpole feature flag service. Built with Go, Fiber v3, GORM, and PostgreSQL.

## Stack

- **Go** 1.26.1
- **[Fiber v3](https://github.com/gofiber/fiber)** — HTTP framework
- **[GORM](https://gorm.io)** — ORM with PostgreSQL driver
- **[golang-jwt/jwt v5](https://github.com/golang-jwt/jwt)** — JWT authentication
- **[swaggo/swag](https://github.com/swaggo/swag)** — Swagger doc generation
- **bcrypt** — Password hashing via `golang.org/x/crypto`

## Project structure

```
src/
├── main.go
├── config/         # Env vars and CLI flag loading
├── database/       # Migrations and seeding
├── models/         # GORM models
├── dal/            # Database access layer
├── controllers/    # Business logic
├── handlers/       # HTTP handlers
├── middleware/     # JWT auth middleware
├── routes/         # Route registration
├── pkg/
│   ├── crypto/     # Password hashing and salt utilities
│   ├── jwtutil/    # JWT claim helpers
│   └── response/   # Response envelope and handler wrapper
└── docs/           # Generated Swagger docs
```

## Configuration

Configuration is loaded from `.env`, environment variables, or CLI flags. CLI flags take the highest precedence.

| Variable | CLI flag | Default | Description |
|---|---|---|---|
| `PORT` | `-port` | `4000` | Server port |
| `JWT_SECRET` | `-jwt-secret` | `change-me` | JWT signing secret |
| `DSN` | `-dsn` | — | PostgreSQL connection string |
| `ENV` | — | — | Set to `dev` to enable request logging and Swagger UI |
| `ALLOW_ORIGIN` | — | `http://localhost:5173` | CORS allowed origin |

Copy `.env.example` to `.env` and fill in the values before running.

### DSN format

```
host=localhost user=postgres password=secret dbname=flagpole port=5432 sslmode=disable
```

## Running

```bash
# Run (requires PostgreSQL)
make run

# Build binary
make build

# Regenerate Swagger docs
make swag
```

On first boot, migrations run automatically and the database is seeded with:
- Roles: `admin`, `editor`, `viewer`
- Organization: `flagpole` (internal — invisible to non-internal users)
- Admin user: `admin@flagpole.dev` — password is randomly generated and printed to stdout

## API

Base path: `/api/v1`

The full API reference — request bodies, response shapes, and error codes — is available via Swagger UI. Start the server with `ENV=dev` and visit `http://localhost:4000/docs`.

## Response format

All endpoints return a consistent JSON envelope:

```json
{ "data": { ... } }
```

Errors:

```json
{ "error": "message" }
```

## Authentication flow

1. `POST /login` returns `{ "data": { "token": "<jwt>" } }`
2. Include the token in subsequent requests: `Authorization: Bearer <token>`
3. The JWT contains `userId`, `email`, `role`, `orgIds`, and `orgNames`
