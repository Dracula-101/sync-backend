package mongo

import (
	"context"
	"errors"
	"fmt"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbConfig struct {
	User        string
	Pwd         string
	Host        string
	Name        string
	MinPoolSize uint16
	MaxPoolSize uint16
	Timeout     time.Duration
}

type Document[T any] interface {
	EnsureIndexes(Database)
	GetValue() *T
	Validate() error
}

type Database interface {
	GetLogger() utils.AppLogger
	GetInstance() *database
	GetClient() *mongo.Client
	GetDatabaseName() string
	Connect()
	Disconnect()
}

type database struct {
	*mongo.Database
	logger  utils.AppLogger
	context context.Context
	config  DbConfig
}

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

func (db *database) GetClient() *mongo.Client {
	client := db.Database.Client()
	if client == nil {
		db.logger.Fatal("Mongo client is nil")
	}
	return client
}

func (db *database) GetDatabaseName() string {
	return db.config.Name
}

func (db *database) GetLogger() utils.AppLogger {
	return db.logger
}

func (db *database) Connect() {
	db.logger.Debug("Connecting to mongo...")
	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s",
		db.config.User, db.config.Pwd, db.config.Host,
	)
	db.logger.Debug("Mongo URI: %s", uri)

	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(db.config.Timeout)
	clientOptions.SetAppName(db.config.Name)
	clientOptions.SetRetryReads(true)
	clientOptions.SetRetryWrites(true)

	clientOptions.SetServerSelectionTimeout(db.config.Timeout)
	clientOptions.SetSocketTimeout(db.config.Timeout)

	clientOptions.SetMinPoolSize(uint64(db.config.MinPoolSize))
	clientOptions.SetMaxPoolSize(uint64(db.config.MaxPoolSize))

	client, err := mongo.Connect(db.context, clientOptions)
	if err != nil {
		db.logger.Fatal("Failed to connect to mongo: %v", err)
	}

	err = client.Ping(db.context, nil)
	if err != nil {
		db.logger.Fatal("Failed to ping mongo: %v", err)
	}
	db.logger.Success("Connected to mongo")
	db.Database = client.Database(db.config.Name)
}

func (db *database) Disconnect() {
	db.logger.Debug("Disconnecting from mongo...")
	err := db.Client().Disconnect(db.context)
	if err != nil {
		db.logger.Fatal("Failed to disconnect from mongo: %v", err)
	}
	db.logger.Success("Disconnected from mongo")
}

func NewObjectID(id string) (primitive.ObjectID, error) {
	i, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = errors.New(id + " is not a valid mongo id")
	}
	return i, err
}
