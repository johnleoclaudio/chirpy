# Chirpy

A Twitter-like social media API built in Go as part of Boot.dev's HTTP server course. Users can create accounts, post chirps (short messages), and interact with a simple web interface.

## Features

- User registration and authentication with JWT tokens
- Post and view chirps (short messages)
- User account management
- Admin metrics and controls
- Web interface for viewing chirps
- Webhook integration with Polka payment system

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL database
- Environment variables configured (see Configuration)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your environment variables in a `.env` file:
   ```
   DB_URL=postgres://username:password@localhost/chirpy?sslmode=disable
   JWT_SECRET=your-secret-key
   POLKA_KEY=your-polka-webhook-key
   PLATFORM=dev
   ```

4. Run database migrations:
   ```bash
   make db_migrate_up
   ```

5. Start the server:
   ```bash
   make run
   ```

The server will start on port 8080.

## API Endpoints

### Authentication
- `POST /api/users` - Create a new user account
- `POST /api/login` - Login and get access/refresh tokens
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token
- `PUT /api/users` - Update user information (requires auth)

### Chirps
- `GET /api/chirps` - List all chirps
- `GET /api/chirps/{id}` - Get a specific chirp
- `POST /api/chirps` - Create a new chirp (requires auth)
- `DELETE /api/chirps/{id}` - Delete a chirp (requires auth, author only)

### Admin
- `GET /admin/metrics` - View API metrics
- `POST /admin/reset` - Reset metrics and database

### Other
- `GET /api/healthz` - Health check endpoint
- `POST /api/polka/webhooks` - Polka payment webhook

## Development

### Build Commands
- `make build` - Build binary to ./bin/out
- `make run` - Run directly with go run main.go
- `go test ./...` - Run all tests

### Database Commands
- `make sql_generate` - Generate Go code from SQL using sqlc
- `make db_migrate_up` - Apply database migrations
- `make db_migrate_down` - Rollback migrations

## Configuration

Required environment variables:
- `DB_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT token signing
- `POLKA_KEY` - API key for Polka webhook verification
- `PLATFORM` - Platform identifier (dev/prod)

## Tech Stack

- **Backend**: Go with standard library HTTP server
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Authentication**: JWT tokens with refresh token support
- **Password Hashing**: Argon2id
- **Migrations**: Goose for database migrations
