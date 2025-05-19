package mongo

import (
	"context"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type QueryBuilder[T any] interface {
	GetLogger() utils.AppLogger
	GetCollection() *mongo.Collection
	SingleQuery() Query[T]
	Query(context context.Context) Query[T]
}

type AggregateBuilder[T any, R any] interface {
	GetLogger() utils.AppLogger
	GetCollection() *mongo.Collection
	SingleAggregate() Aggregator[T, R]
	Aggregate(context context.Context) Aggregator[T, R]
}

type queryBuilder[T any] struct {
	logger     utils.AppLogger
	collection *mongo.Collection
	timeout    time.Duration
}

type aggregateBuilder[T any, R any] struct {
	logger     utils.AppLogger
	collection *mongo.Collection
	timeout    time.Duration
}

func (c *queryBuilder[T]) GetCollection() *mongo.Collection {
	return c.collection
}

func (c *queryBuilder[T]) GetLogger() utils.AppLogger {
	return c.logger
}

func (c *queryBuilder[T]) SingleQuery() Query[T] {
	return newSingleQuery[T](c.logger, c.collection, c.timeout)
}

func (c *queryBuilder[T]) Query(context context.Context) Query[T] {
	return newQuery[T](c.logger, context, c.collection)
}

func (a *aggregateBuilder[T, R]) GetCollection() *mongo.Collection {
	return a.collection
}

func (a *aggregateBuilder[T, R]) GetLogger() utils.AppLogger {
	return a.logger
}

func (a *aggregateBuilder[T, R]) SingleAggregate() Aggregator[T, R] {
	return newSingleAggregator[T, R](a.logger, a.collection, a.timeout)
}

func (a *aggregateBuilder[T, R]) Aggregate(context context.Context) Aggregator[T, R] {
	return newAggregator[T, R](a.logger, context, a.collection)
}

func NewQueryBuilder[T any](db Database, collectionName string) QueryBuilder[T] {
	return &queryBuilder[T]{
		collection: db.GetInstance().Collection(collectionName),
		timeout:    time.Minute * 5,
		logger:     db.GetLogger(),
	}
}

func NewAggregateBuilder[T any, R any](db Database, collectionName string) AggregateBuilder[T, R] {
	return &aggregateBuilder[T, R]{
		collection: db.GetInstance().Collection(collectionName),
		timeout:    time.Minute * 5,
		logger:     db.GetLogger(),
	}
}

type TransactionBuilder interface {
	GetLogger() utils.AppLogger
	GetDatabase() string
	GetClient() *mongo.Client
	GetTransaction(timeout time.Duration) Transaction
}

type transactionBuilder struct {
	logger   utils.AppLogger
	client   *mongo.Client
	database string
}

func (t *transactionBuilder) GetLogger() utils.AppLogger {
	return t.logger
}

func (t *transactionBuilder) GetDatabase() string {
	return t.database
}

func (t *transactionBuilder) GetClient() *mongo.Client {
	return t.client
}

func (t *transactionBuilder) GetTransaction(timeout time.Duration) Transaction {
	return newTransaction(t.logger, t.client, t.database, timeout)
}

func NewTransactionBuilder(db Database) TransactionBuilder {
	return &transactionBuilder{
		logger:   db.GetLogger(),
		client:   db.GetClient(),
		database: db.GetDatabaseName(),
	}
}
