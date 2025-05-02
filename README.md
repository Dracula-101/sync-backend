# Sync Backend

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24.2-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)
![Last Updated](https://img.shields.io/badge/updated-May%202025-brightgreen.svg)

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
  - MongoDB for primary data storage
  - PostgreSQL for IP geolocation data
  - Redis for caching and rate limiting
  - Structured logging system
  - Environment-based configuration management

## Getting Started

### Prerequisites

- Go 1.24.2 or later
- MongoDB
- PostgreSQL
- Redis

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
go run main.go
```

## Project Structure

```
sync-backend/
├── api/                          # API components organized by domain
│   ├── auth/                     # Authentication API
│   │   ├── controller.go
│   │   ├── service.go
│   │   ├── dto/                  # Data Transfer Objects
│   │   ├── middleware/           # Auth middlewares
│   │   └── model/                # Auth domain models
│   ├── common/                   # Common services (location, sessions, etc)
│   ├── community/                # Community related functionality
│   ├── post/                     # Post related functionality
│   └── user/                     # User related functionality
├── arch/                         # Core architecture components
│   ├── application/              # Application bootstrapping
│   ├── common/                   # Common utilities
│   ├── config/                   # Configuration management
│   ├── dto/                      # Common DTOs
│   ├── middleware/               # Core middlewares
│   ├── mongo/                    # MongoDB integration
│   ├── network/                  # HTTP networking components
│   ├── postgres/                 # PostgreSQL integration
│   ├── redis/                    # Redis integration
│   └── utils/                    # Architectural utilities
├── configs/                      # Configuration files
├── keys/                         # Cryptographic keys
├── scripts/                      # Utility scripts
├── test/                         # Test helpers
├── utils/                        # General utility functions
└── .tools/                       # Development tooling
```

## Configuration

The application uses a combination of YAML configuration files and environment variables:

- `configs/app.yaml`: General application settings
- `configs/auth.yaml`: Authentication settings
- `configs/db.yaml`: Database configuration
- `.env`: Environment-specific variables

## Development


This creates the boilerplate for a new feature with:
1. Models (MongoDB schemas)
2. DTOs (Data Transfer Objects)
3. Service layer
4. Controller with basic CRUD endpoints

### Manual Creation

1. Create feature folder in `api/{feature_name}`
2. Implement model, service, and controller
3. Register the controller in `arch/application/module.go`

## Testing

```bash
# Run all tests
go test ./...

# Run specific test package
go test ./api/auth/...
```

## Key Components

### Modular Architecture
- Feature-based organization in the `api/` directory
- Core components in the `arch/` directory
- Clear separation of concerns between layers

### Database Integration
- MongoDB for main application data via `arch/mongo` package
- PostgreSQL for IP geolocation via `arch/postgres` package
- Redis for caching and rate-limiting via `arch/redis` package

### Configuration Management
- YAML-based configuration with [Viper](https://github.com/spf13/viper)
- Environment variable interpolation in config files
- Strongly-typed configuration objects

### Networking Layer
- [Gin Web Framework](https://github.com/gin-gonic/gin) for HTTP routing and middleware
- Consistent error handling and response formatting
- Content validation with go-playground/validator

### Authentication
- JWT-based authentication via golang-jwt/jwt
- Session management with device tracking
- Role-based access control

### Development Tools
- Live reloading with [Air](https://github.com/cosmtrek/air)
- Environment variable support with `.env` files
- Code generation for new API features

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
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- [Redis Go Client](https://github.com/redis/go-redis)
- [Viper](https://github.com/spf13/viper) for configuration
- [Validator](https://github.com/go-playground/validator) for data validation

## Contact

Project Link: [https://github.com/yourusername/sync-backend](https://github.com/yourusername/sync-backend)

*Last updated: May 3, 2025*