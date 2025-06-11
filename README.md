# Sync Backend

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.24.2-00ADD8.svg)](https://go.dev/)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/Dracula-101/sync-backend)
[![Last Updated](https://img.shields.io/badge/updated-June%2003%2C%202025-informational.svg)](https://github.com/Dracula-101/sync-backend)

A Go-powered backend service for the Sync social media platform - built for performance, scalability, and developer friendliness. Handles authentication, content management, social interactions, and user data with robust security measures.

## 🚀 Quick Start

To get started quickly, follow these steps:
1. **Clone the repository:**
  ```bash
  git clone https://github.com/Dracula-101/sync-backend
  cd sync-backend
  ```

2. **Install dependencies:**
  ```bash
  go mod tidy
  ```

3. **Set up your environment variables:**
Create a `.env` file in the project root based on the provided `.env.example`:
  ```bash
  cp .env.example .env
  ```

4. **Configure the env secrets:**
Edit the `.env` file with your specific configuration, including database connections, JWT secrets, and Redis settings. Kindly refer to the [env setup](docs/ENV-SETUP.md) for details on required variables.
  ```bash
  # Example .env content
  HOST=localhost
  PORT=8080
  ENV=development
  LOG_LEVEL=info
  ```

5. **Run the application:**
  ```bash
  go run main.go
  ```

> [!TIP]
>  Import the postman collection from `/postman` to test the API endpoints.

## ✨ Features

- **Authentication System**
  - Secure email/password login with bcrypt hashing
  - JWT authentication with refresh tokens
  - Google OAuth integration
  - Multi-device session management
  - Password reset via email

- **Social Features**
  - User profiles with follow relationships
  - Media-rich posts with text, images, and videos
  - Community creation and moderation
  - Threaded comments with reactions
  - Content tagging and search

- **Performance Optimizations**
  - MongoDB for core social content
  - Redis for caching and rate limiting
  - PostgreSQL for location services
  - Optimized query patterns and indexing

- **API Design**
  - RESTful endpoints with v1 namespace
  - Consistent error responses
  - Rate limiting by IP and user
  - Input validation with detailed feedback

## 🛠️ Requirements

- **Go 1.24.2+** - Core language runtime
- **MongoDB 6.0+** - Primary database for social content and user data
- **Redis 7.0+** - Caching, rate limiting, and session management
- **PostgreSQL 14+** - Geographic data and advanced analytics
- **Air** - For hot-reload during development

## 🏃‍♂️ Development

**With hot-reload:**
```bash
go install github.com/cosmtrek/air@latest
air
```

**Manual run:**
```bash
go run main.go
```

**Project structure:**
```
sync-backend/
├── api/                  # API components organized by domain
│   ├── auth/             # Authentication (login, signup, OAuth)
│   ├── comment/          # Comment functionality and reactions
│   ├── common/           # Shared services (location, analytics)
│   ├── community/        # Community creation and management
│   ├── post/             # Post creation and interaction
│   ├── user/             # User profiles and relationships
│   └── system/           # System-wide operations
├── arch/                 # Core architecture components
│   ├── application/      # App bootstrapping and configuration
│   ├── mongo/            # MongoDB integration layer
│   ├── postgres/         # PostgreSQL integration
│   ├── redis/            # Redis caching and storage
│   ├── network/          # HTTP networking layer
│   └── middleware/       # Global middleware (errors, rate limiting)
├── configs/              # Configuration files (YAML)
├── scripts/              # Deployment and maintenance scripts
├── seed/                 # Seed data for initial setup
├── test/                 # Testing utilities
├── uploads/              # File upload storage location
└── utils/                # Helper utilities and tools
```

## 🏗️ Architecture

- **Domain-Driven Design** - Business logic organized by domain
- **Clean Architecture** - Separation between controllers, services, and models
- **Modular Components** - Controllers, Services, Models, and DTOs
- **Data Storage** - MongoDB (documents), Redis (caching), PostgreSQL (geo)
- **Networking** - Gin framework, custom middleware, request validation
- **Authentication** - JWT tokens with multi-device session management

For details on implementation patterns and code examples, see [ARCHITECTURE.md](docs/ARCHITECTURE.md).

## 🤝 Contributing

Contributions are welcome! Here's how to contribute to Sync Backend:

1. **Fork** the repository
2. **Clone** your fork (`git clone https://github.com/Dracula-101/sync-backend.git`)
3. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
4. **Make** your changes
   - Follow the existing code style
   - Add tests for new functionality
   - Ensure all tests pass (`go test ./...`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Document exported functions
- Write tests for new functionality

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin) - Web framework
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - MongoDB driver
- [Redis Go Client](https://github.com/redis/go-redis) - Redis client
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Validator](https://github.com/go-playground/validator) - Request validation
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT implementation
- [Air](https://github.com/cosmtrek/air) - Live reload for development

## 📬 Contact

For questions, issues, or feedback, please contact us via - [Email](mailto:pratikpujari1000@gmail.com), [Create an issue](https://github.com/Dracula-101/sync-backend/issues)

---

*Last updated: May 21, 2025*
