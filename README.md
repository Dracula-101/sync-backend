# Sync Backend

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

A robust backend service for the "Sync" social media platform, providing secure authentication, scalable API endpoints, and performance-optimized services.

## Features

- **Modern Authentication System**
  - Email/password login with secure password hashing
  - JWT-based authentication
  - OAuth integration (Google support)
  - Session management with revocation capabilities
  
- **API Architecture**
  - RESTful API design with versioning
  - Rate limiting for API protection
  - CORS configuration for frontend compatibility
  
- **Robust Infrastructure**
  - Dependency injection using Uber's FX
  - Structured logging with Zap
  - Environment-based configuration management
  - Command-line interface with Cobra

- **Security-First Approach**
  - Rate limiting on sensitive endpoints
  - Password policy enforcement
  - CSRF protection
  - Sanitized logging (removing sensitive data)

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for local development)

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/sync-backend.git
cd sync-backend
```

2. Copy the example environment file
```bash
cp .env.example .env
```

3. Install dependencies
```bash
go mod download
```

### Running the Application

#### Development Mode

For development with hot-reload using Air:

```bash
air
```

#### Manual Run

```bash
go run main.go app:serve
```

#### Using Docker

```bash
docker-compose up -d
```

## Project Structure

```
sync-backend/
├── cmd/                          # Application entry points
├── configs/                      # Configuration files
├── internal/                     # Internal application code
│   ├── api/                     # API handlers and routes
│   ├── application/             # Application services
│   ├── domain/                  # Domain models and interfaces
│   ├── infrastructure/          # Infrastructure implementations
│   ├── server/                  # Server configuration
│   └── util/                    # Utility functions
├── pkg/                          # Reusable packages
│   ├── console/                 # Command line utilities
│   └── logger/                  # Logging utilities
├── migrations/                   # Database migrations
├── scripts/                      # Utility scripts
├── test/                         # Test files
└── docker/                       # Docker configuration
```

## Configuration

The application uses a combination of YAML configuration files and environment variables:

- `configs/app.yaml`: General application settings
- `configs/auth.yaml`: Authentication settings
- `configs/log.yaml`: Logging configuration
- `.env`: Environment-specific variables

## Development

### Adding a New API Endpoint

1. Create a new route handler in `internal/api/handlers/`
2. Add the route definition in `internal/api/route/`
3. Register the route in the main router

### Command Line Interface

The application uses Cobra for CLI commands. The main command structure:

```bash
# Run the server
go run main.go app:serve

# Additional commands can be added in internal/application/app.go
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific test package
go test ./internal/application/auth/...
```

## Key Components

### Command-based Architecture
- The application uses [Cobra](https://github.com/spf13/cobra) for command-line interface
- Commands are defined in `internal/application/app.go`
- The main command (`app:serve`) starts the HTTP server

### Dependency Injection
- Uses [Uber's fx](https://github.com/uber-go/fx) for dependency management
- Services are registered in `internal/application/modules.go`

### Configuration Management
- YAML-based configuration with [Viper](https://github.com/spf13/viper)
- Environment variable interpolation in config files
- Separate config files for different concerns (app, auth, log)

### Logging
- Structured logging with [Zap](https://github.com/uber-go/zap)
- Custom logger implementation in `pkg/logger/logger.go`

### HTTP Server
- Gin web framework for routing
- Middleware support (rate limiting, CORS)
- Route grouping by API version

### Authentication
- Basic authentication routes defined in `internal/api/route/auth_routes.go`
- JWT-based authentication (configured but not fully implemented)

### Development Tools
- Live reloading with [Air](https://github.com/cosmtrek/air)
- Environment variable support with `.env` files

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Uber's FX](https://github.com/uber-go/fx) for dependency injection
- [Cobra](https://github.com/spf13/cobra) for CLI commands
- [Viper](https://github.com/spf13/viper) for configuration
- [Zap](https://github.com/uber-go/zap) for structured logging

## Contact

Project Link: [https://github.com/yourusername/sync-backend](https://github.com/yourusername/sync-backend)