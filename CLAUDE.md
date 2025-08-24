# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based GraphQL API service called "scraper-api" that provides anime metadata management through integration with TheTVDB API. The service uses Federation GraphQL schema and includes authentication-protected endpoints.

## Architecture

### Core Components
- **CLI Application**: Built with Cobra CLI framework, main commands in `internal/commands/`
- **GraphQL Server**: Uses gqlgen with Federation support, schema defined in `graph/schema.graphqls`
- **Database Layer**: PostgreSQL with GORM, migrations in `db/migrations/`
- **HTTP Layer**: Chi router with middleware for logging, CORS, and request info
- **Services**: 
  - TheTVDB API integration (`internal/services/thetvdb_api/`)
  - Link management (`internal/services/link_service/`)
  - Episode and anime services
- **Metrics**: Custom metrics library in `metrics_lib/` with DataDog and Prometheus support
- **Message Queue**: Apache Pulsar integration for async processing

### Key Directories
- `cmd/main.go`: Application entry point
- `internal/commands/`: CLI commands (serve, migrate, sync, etc.)
- `graph/`: GraphQL schema, resolvers, and generated code
- `internal/db/repositories/`: Database entities and repositories
- `internal/services/`: Business logic services
- `http/`: HTTP server and handlers
- `config/`: Configuration management

## Development Commands

### Build and Run
```bash
# Build the application
go build -o scraper-api ./cmd/main.go

# Run the server
go run cmd/main.go serve

# Using Docker
docker build -t scraper-api .
docker run -p 3000:3000 scraper-api
```

### Database Operations
```bash
# Run migrations
go run cmd/main.go migrate up

# Create new migration
make create-migration name=migration_name

# Direct migration command
go run github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir db/migrations MIGRATION_NAME
```

### Code Generation
```bash
# Generate GraphQL resolvers and types
make gql
# or directly:
go run github.com/99designs/gqlgen generate

# Generate mocks
make mocks

# Generate all (mocks + GraphQL)
make generate
```

### Testing
```bash
# Run tests
go test ./...

# Run specific test file
go test ./internal/services/thetvdb_api/

# Run tests with coverage
go test -cover ./...
```

### Development Environment
```bash
# Start dependencies (MySQL, Redis)
docker-compose up

# The service expects:
# - MySQL on port 3306
# - Redis on port 6379
# - Configuration in config/config.dev.json
```

## Configuration

The application uses JSON configuration files in the `config/` directory. Configuration is loaded via the `jinzhu/configor` package.

## Authentication

The GraphQL schema includes `@Authenticated` directives on protected endpoints. Authentication middleware is implemented in the HTTP layer.

## Database Schema

Key entities:
- `thetvdb_link`: Links between anime IDs and TheTVDB IDs
- `anime`: Anime metadata
- `anime_episode`: Episode information

Migrations are managed through `golang-migrate/migrate/v4` and stored in `db/migrations/`.

## Local Development Notes

- The application uses CGO for some dependencies (librdkafka)
- Metrics are collected via custom middleware and sent to DataDog
- GraphQL Federation is enabled for service composition
- The service exposes health check endpoints through the HTTP layer