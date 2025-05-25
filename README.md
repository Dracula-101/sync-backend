# Sync Backend

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.24.2-00ADD8.svg)](https://go.dev/)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/Dracula-101/sync-backend)
[![Last Updated](https://img.shields.io/badge/updated-May%2021%2C%202025-informational.svg)](https://github.com/Dracula-101/sync-backend)

A Go-powered backend service for the Sync social media platform - built for performance, scalability, and developer friendliness. Handles authentication, content management, social interactions, and user data with robust security measures.

## 🚀 Quick Start

```bash
# Clone repository
git clone https://github.com/Dracula-101/sync-backend.git
cd sync-backend

# Set up configuration (update database credentials in .env)
cp .env.example .env

# Install dependencies
go mod download

# Run the server
go run main.go
```

## ✨ Features

- **Authentication System**
  - Email/password login with secure password hashing
  - JWT-based token authentication
  - Google OAuth integration
  - Session management with device tracking and revocation
  - Password reset functionality

- **Social Networking Core**
  - User profiles and relationships
  - Posts with rich media support
  - Communities with moderation
  - Nested comments and reactions
  - Content tagging system

- **Performance Infrastructure**
  - MongoDB for document storage (user data, posts, communities)
  - Redis for caching and rate limiting
  - PostgreSQL for geolocation services
  - Optimized query patterns

- **API Design**
  - RESTful endpoints with versioning
  - Consistent error handling
  - Rate limiting protection
  - Comprehensive request validation
  - CORS configuration for frontend compatibility

## 🛠️ Requirements

- **Go 1.24.2+** - Core language runtime
- **MongoDB 6.0+** - Primary database for social content and user data
- **Redis 7.0+** - Caching, rate limiting, and session management
- **PostgreSQL 14+** - Geographic data and advanced analytics
- **Air** - For hot-reload during development (optional)

## 🏃‍♂️ Development

**With hot-reload:**
```bash
# Install Air first if not installed
go install github.com/cosmtrek/air@latest

# Run with automatic reloading
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

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run auth tests only
go test ./api/auth/...

# Run with coverage report
go test -cover ./...
```

## 🔧 Configuration

The application uses a combination of YAML files and environment variables:

- **configs/app.yaml** - Core application settings, logging, server ports
- **configs/auth.yaml** - Authentication settings, JWT secrets, OAuth credentials
- **configs/db.yaml** - Database connection strings and settings
- **.env** - Environment-specific variables and secrets

Example `.env` file structure:
```
# Server
HOST=localhost
PORT=8080
ENV=development
LOG_LEVEL=info

JWT_SECRET=<your_jwt_secret>

# Database connections
DB_HOST=
DB_NAME=
DB_USER=
DB_PASSWORD=

# IP Geolocation DB
IP_DB_HOST=
IP_DB_PORT=
IP_DB_USER=
IP_DB_PASSWORD=
IP_DB_NAME=

# Redis connections
REDIS_HOST=
REDIS_PORT=
REDIS_PASSWORD=
REDIS_DB=

# Cloudinary for uploading media
CLOUDINARY_CLOUD_NAME=
CLOUDINARY_API_KEY=
CLOUDINARY_API_SECRET=
```

## 🏗️ Architecture

- **Domain-Driven Design** - Business logic organized by domain in the `api/` directory
- **Clean Architecture** - Clear separation between controllers, services, and data models
- **Modular Components**:
  - **Controllers** - Handle HTTP requests and responses
  - **Services** - Contain business logic and orchestrate operations
  - **Models** - Define data structures and MongoDB schemas
  - **DTOs** - Handle data transfer between layers

- **Data Storage**:
  - **MongoDB** - Document storage for users, posts, communities, comments
  - **Redis** - Session storage, caching, rate limiting
  - **PostgreSQL** - Geolocation services and relational data

- **Networking**:
  - **[Gin Web Framework](https://github.com/gin-gonic/gin)** - Fast HTTP routing and middleware
  - **Custom Middleware** - Error handling, logging, authentication
  - **Validation** - Request validation with [go-playground/validator](https://github.com/go-playground/validator)

- **Authentication**:
  - **JWT** - Stateless authentication with [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
  - **Session Management** - Multi-device login with revocation

## 🛠️ Adding New Features

The codebase follows a consistent pattern to make adding new features straightforward:

1. **Create feature directory** in `api/{feature_name}/`
2. **Implement these files**:
   - `controller.go` - HTTP handlers and routes
   - `service.go` - Business logic
   - `error.go` - Feature-specific error definitions
   - `dto/` - Request/response data structures
   - `model/` - Data models for MongoDB (if needed)
   - `middleware/` - Feature-specific middleware (if needed)
3. **Register controller** in `arch/application/module.go`

**Example of adding a new "notification" feature:**

```go
// api/notification/service.go

type NotificationService interface {
  CreateNotification(userID string, notification Notification) error
  GetNotifications(userID string) ([]Notification, error)
  DeleteNotification(notificationID string) error
}

type notificationService struct {
  logger utils.AppLogger
  // other dependencies...
}

func NewNotificationService() *notificationService {
  return &notificationService{
    logger: utils.NewServiceLogger("notification"),
    // other initializations...
  }
}

// Implement methods for creating, retrieving, and deleting notifications
```


```go
// api/notification/controller.go

package notification

import (
  "sync-backend/arch/network"
  "sync-backend/arch/common"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type notificationController struct {
  	logger utils.AppLogger
    network.BaseController
    common.ContextPayload
  	authProvider     network.AuthenticationProvider
    service          *NotificationService
    
}

func NewNotificationController(
  authProvider network.AuthenticationProvider,
  service *NotificationService,
) *network.Controller {
    return &notificationController{
      logger:           utils.NewServiceLogger("notification"),
      BaseController:   network.NewBaseController("/api/v1/notification", authProvider),
      ContextPayload:   common.NewContextPayload(),
      authProvider:     authProvider,
      service:          service,
    }
}

func (c *controller) MountRoutes(group *gin.RouterGroup) {
    group.GET("/notifications", c.GetNotifications)
    group.POST("/notifications", c.CreateNotification)
    group.DELETE("/notifications/:id", c.DeleteNotification)
}

// Add handler implementations...
```

Then register in `arch/application/module.go`:

```go
// Add to Controllers function
func (m *appModule) Controllers() []network.Controller {
  // Other controllers...
  notificationController := notification.NewNotificationController(
    m.authProvider,
    m.notificationService,
  )
}

func NewAppModule(
  // Other dependencies...
) Module {
  // Other initializations...
  notificationService := notification.NewNotificationService()
  return &appModule{
    // Other dependencies...
    notificationService: notificationService,
  }
}

```

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

Project Link: [https://github.com/Dracula-101/sync-backend](https://github.com/Dracula-101/sync-backend)

---

*Last updated: May 21, 2025*