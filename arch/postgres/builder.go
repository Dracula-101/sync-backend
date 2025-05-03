package postgres

import (
	"context"
	"database/sql"
	"sync-backend/utils"
	"time"
)

// QueryBuilder interface for PostgreSQL
type QueryBuilder[T any] interface {
	GetLogger() utils.AppLogger
	GetDB() *sql.DB
	SingleQuery() Query[T]
	Query(context context.Context) Query[T]
}

type queryBuilder[T any] struct {
	logger  utils.AppLogger
	db      *sql.DB
	timeout time.Duration
}

func (b *queryBuilder[T]) GetDB() *sql.DB {
	return b.db
}

func (b *queryBuilder[T]) GetLogger() utils.AppLogger {
	return b.logger
}

func (b *queryBuilder[T]) SingleQuery() Query[T] {
	return newSingleQuery[T](b.logger, b.db, b.timeout)
}

func (b *queryBuilder[T]) Query(context context.Context) Query[T] {
	return newQuery[T](b.logger, context, b.db)
}

// NewQueryBuilder creates a new PostgreSQL query builder
func NewQueryBuilder[T any](db Database) QueryBuilder[T] {
	return &queryBuilder[T]{
		db:      db.GetDB(),
		timeout: time.Minute * 5,
		logger:  db.GetLogger(),
	}
}
