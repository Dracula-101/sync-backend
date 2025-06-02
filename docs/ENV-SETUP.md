# Environment Configuration Guide

This document provides detailed information on configuring the environment for the Sync Backend application. The application uses a combination of YAML configuration files and environment variables to manage various aspects of its behavior.

## Configuration Files Overview

Sync Backend uses three main configuration files, all located in the `configs/` directory:

1. **app.yaml** - Core application settings
2. **auth.yaml** - Authentication settings
3. **db.yaml** - Database connection settings

These files use environment variable interpolation with `${VARIABLE_NAME}` syntax to load sensitive or environment-specific values from the `.env` file.

## Environment Variables

The `.env` file contains all sensitive information and environment-specific configurations. Create this file by copying the example:

```bash
copy .env.example .env
```

### Required Environment Variables

#### Server Configuration
This section defines the basic server settings:

- `HOST` - The host address to bind the server (e.g., `localhost` or `0.0.0.0`)
- `PORT` - The port number for the server (e.g., `8080`)
- `ENV` - Environment name (`development`, `staging`, `production`)
- `LOG_LEVEL` - Logging level (`debug`, `info`, `warn`, `error`)

#### Security
- `JWT_SECRET` - Secret key for JWT token generation and validation

#### MongoDB Connection
This section defines the MongoDB connection settings, usually a hosted [Mongo Altas instance](https://www.mongodb.com/) or a local MongoDB server:
- `DB_HOST` - MongoDB host address (e.g., `localhost` or connection string)
- `DB_NAME` - Database name (e.g., `sync`)
- `DB_USER` - MongoDB username
- `DB_PASSWORD` - MongoDB password

#### PostgreSQL for Geolocation
This section defines the PostgreSQL connection settings for IP geolocation data:
- `IP_DB_HOST` - PostgreSQL host address
- `IP_DB_PORT` - PostgreSQL port (typically `5432`)
- `IP_DB_NAME` - Database name for geolocation data
- `IP_DB_USER` - PostgreSQL username
- `IP_DB_PASSWORD` - PostgreSQL password

Use this [tutorial](https://dev.maxmind.com/geoip/importing-databases/postgresql/) to set up PostgreSQL for geolocation


#### Redis Configuration
- `REDIS_HOST` - Redis host address
- `REDIS_PORT` - Redis port (typically `6379`)
- `REDIS_PASSWORD` - Redis password
- `REDIS_DB` - Redis database index (e.g., `0`)

#### Media Storage
Apply for a [Cloudinary](https://cloudinary.com/) account to manage media uploads:
- `CLOUDINARY_CLOUD_NAME` - Cloudinary cloud name
- `CLOUDINARY_API_KEY` - Cloudinary API key
- `CLOUDINARY_API_SECRET` - Cloudinary API secret

### Environment-Specific Configurations

#### Production Environment
```
HOST=0.0.0.0
PORT=8080
ENV=production
LOG_LEVEL=info

JWT_SECRET=your_secure_random_string_here

DB_HOST=mongodb://mongodb.your-domain.com
DB_PORT=27017
DB_NAME=sync_prod
DB_USER=prod_user
DB_PASSWORD=strong_password_here

IP_DB_HOST=postgres.your-domain.com
IP_DB_PORT=5432
IP_DB_NAME=geo_prod
IP_DB_USER=geo_user
IP_DB_PASSWORD=strong_postgres_password

REDIS_HOST=redis.your-domain.com
REDIS_PORT=6379
REDIS_PASSWORD=strong_redis_password
REDIS_DB=0

CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

## Environment Validation

On startup, the application validates required environment variables. If any required variable is missing, the application will fail to start with an appropriate error message.

## Secrets Management

For local development, secrets are stored in the `.env` file, which should **never** be committed to version control.

For production deployments, consider using:
- Environment variables set through your deployment platform
- A secure vault system like HashiCorp Vault
- Kubernetes Secrets
- Cloud provider secret management services (e.g., AWS Secrets Manager, GCP Secret Manager)

## Common Configuration Issues

### Database Connection Failures
- Check if the database server is running
- Verify the connection credentials in `.env` file
- Ensure network connectivity to the database server

### JWT Authentication Issues
- Verify that `JWT_SECRET` is properly set
- Check that the token expiry times are appropriate

### Redis Connection Problems
- Ensure Redis server is running
- Verify password configuration (if any)
- Check Redis server reachability

---

*Last updated: June 2, 2025*