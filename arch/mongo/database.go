package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	GetInstance() *database
	Connect()
	Disconnect()
}

type database struct {
	*mongo.Database
	context context.Context
	config  DbConfig
}

func NewDatabase(ctx context.Context, config DbConfig) Database {
	db := database{
		context: ctx,
		config:  config,
	}
	return &db
}

func (db *database) GetInstance() *database {
	return db
}

func (db *database) Connect() {
	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s",
		db.config.User, db.config.Pwd, db.config.Host,
	)
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(db.config.Timeout)
	clientOptions.SetAppName(db.config.Name)
	clientOptions.SetRetryReads(true)
	clientOptions.SetRetryWrites(true)

	clientOptions.SetServerSelectionTimeout(db.config.Timeout)
	clientOptions.SetSocketTimeout(db.config.Timeout)

	clientOptions.SetMinPoolSize(uint64(db.config.MinPoolSize))
	clientOptions.SetMaxPoolSize(uint64(db.config.MaxPoolSize))

	fmt.Println("connecting mongo...")
	client, err := mongo.Connect(db.context, clientOptions)
	if err != nil {
		log.Fatal("connection to mongo failed!: ", err)
	}

	err = client.Ping(db.context, nil)
	if err != nil {
		log.Panic("pinging to mongo failed!: ", err)
	}
	fmt.Println("connected to mongo!")

	db.Database = client.Database(db.config.Name)
}

func (db *database) Disconnect() {
	fmt.Println("disconnecting mongo...")
	err := db.Client().Disconnect(db.context)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("disconnected mongo")
}

func NewObjectID(id string) (primitive.ObjectID, error) {
	i, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = errors.New(id + " is not a valid mongo id")
	}
	return i, err
}
