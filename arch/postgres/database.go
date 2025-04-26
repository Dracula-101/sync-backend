package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync-backend/utils"
	"time"

	_ "github.com/lib/pq"
)

// DbConfig holds the configuration for PostgreSQL connection
type DbConfig struct {
	User         string
	Pwd          string
	Host         string
	Port         string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// Document interface for PostgreSQL models
type Document[T any] interface {
	EnsureSchema(Database)
	GetValue() *T
	Validate() error
}

// Database interface for PostgreSQL database operations
type Database interface {
	GetLogger() utils.AppLogger
	GetInstance() *database
	GetDB() *sql.DB
	GetDatabaseName() string
	Connect()
	Disconnect()
}

// database struct implements Database interface
type database struct {
	*sql.DB
	logger  utils.AppLogger
	context context.Context
	config  DbConfig
}

// NewDatabase creates a new PostgreSQL database instance
func NewDatabase(ctx context.Context, logger utils.AppLogger, config DbConfig) Database {
	db := database{
		context: ctx,
		logger:  logger,
		config:  config,
	}
	return &db
}

func (db *database) GetInstance() *database {
	return db
}

func (db *database) GetDB() *sql.DB {
	if db.DB == nil {
		db.logger.Fatal("PostgreSQL DB is nil")
	}
	return db.DB
}

func (db *database) GetDatabaseName() string {
	return db.config.Name
}

func (db *database) GetLogger() utils.AppLogger {
	return db.logger
}

func (db *database) Connect() {
	db.logger.Debug("Connecting to PostgreSQL")

	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.config.User, db.config.Pwd, db.config.Host,
		db.config.Port, db.config.Name, db.config.SSLMode,
	)
	db.logger.Debug("PostgreSQL connection string: %s", connectionString)

	var err error
	db.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		db.logger.Fatal("Failed to create PostgreSQL connection: %v", err)
	}

	db.DB.SetMaxOpenConns(db.config.MaxOpenConns)
	db.DB.SetMaxIdleConns(db.config.MaxIdleConns)
	db.DB.SetConnMaxLifetime(db.config.MaxLifetime)

	// Test the connection
	ctx, cancel := context.WithTimeout(db.context, 5*time.Second)
	defer cancel()

	err = db.DB.PingContext(ctx)
	if err != nil {
		db.logger.Fatal("Failed to ping PostgreSQL: %v", err)
	}
	db.logger.Success("Connected to PostgreSQL")
}

func (db *database) Disconnect() {
	db.logger.Debug("Disconnecting from PostgreSQL")
	if db.DB != nil {
		err := db.DB.Close()
		if err != nil {
			db.logger.Error("Failed to disconnect from PostgreSQL: %v", err)
		} else {
			db.logger.Success("Disconnected from PostgreSQL")
		}
	}
}
