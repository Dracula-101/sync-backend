package mongo

import (
	"context"
	"fmt"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Transaction interface {
	Start() error
	PerformTransaction(fn func() error) error
	Commit() error
	Abort() error
	WithTimeout(timeout time.Duration) Transaction
	GetContext() context.Context
	GetCollection(name string) *mongo.Collection
	Close()
}

type transaction struct {
	logger     utils.AppLogger
	client     *mongo.Client
	session    mongo.Session
	database   string
	context    context.Context
	cancel     context.CancelFunc
	hasTimeout bool
}

func newTransaction(logger utils.AppLogger, client *mongo.Client, database string, timeout time.Duration) Transaction {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &transaction{
		logger:     logger,
		client:     client,
		database:   database,
		context:    ctx,
		cancel:     cancel,
		hasTimeout: true,
	}
}

func (t *transaction) WithTimeout(timeout time.Duration) Transaction {
	if t.hasTimeout && t.cancel != nil {
		t.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.context = ctx
	t.cancel = cancel
	t.hasTimeout = true
	return t
}

func (t *transaction) Start() error {
	t.logger.Info("[ MONGO ] - Starting transaction")
	var err error
	t.session, err = t.client.StartSession()
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start session: %v", err)
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = t.session.Client().Ping(t.context, readpref.Primary())
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to ping primary: %v", err)
		t.session.EndSession(t.context)
		return fmt.Errorf("failed to ping primary: %w", err)
	}

	err = t.session.StartTransaction(options.Transaction())
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start transaction: %v", err)
		t.session.EndSession(t.context)
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction started successfully")
	return nil
}

func (t *transaction) PerformTransaction(fn func() error) error {
	t.logger.Debug("[ MONGO ] - Performing transaction")
	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot perform transaction: no active session")
		return fmt.Errorf("cannot perform transaction: no active session")
	}

	_, err := t.session.WithTransaction(t.context, func(sessCtx mongo.SessionContext) (interface{}, error) {
		t.logger.Debug("[ MONGO ] - Executing transaction function")
		err := fn()
		if err != nil {
			t.logger.Error("[ MONGO ] - Transaction function failed: %v", err)
			return nil, err
		}
		t.logger.Debug("[ MONGO ] - Transaction function executed successfully")
		return nil, nil
	})
	if err != nil {
		t.logger.Error("[ MONGO ] - Transaction failed: %v", err)
		t.session.AbortTransaction(t.context)
		t.session.EndSession(t.context)
		return fmt.Errorf("transaction failed: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction performed successfully")
	return nil
}

func (t *transaction) Commit() error {
	t.logger.Info("[ MONGO ] - Committing transaction")
	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot commit: no active session")
		return fmt.Errorf("cannot commit: no active session")
	}

	err := t.session.CommitTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.session.EndSession(t.context)
	t.logger.Info("[ MONGO ] - Transaction committed successfully")
	return nil
}

func (t *transaction) Abort() error {
	t.logger.Info("[ MONGO ] - Aborting transaction")
	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot abort: no active session")
		return fmt.Errorf("cannot abort: no active session")
	}

	err := t.session.AbortTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to abort transaction: %v", err)
		return fmt.Errorf("failed to abort transaction: %w", err)
	}

	t.session.EndSession(t.context)
	t.logger.Info("[ MONGO ] - Transaction aborted successfully")
	return nil
}

func (t *transaction) GetContext() context.Context {
	return t.context
}

func (t *transaction) GetCollection(name string) *mongo.Collection {
	return t.client.Database(t.database).Collection(name)
}

func (t *transaction) Close() {
	if t.hasTimeout && t.cancel != nil {
		t.cancel()
	}
	if t.session != nil {
		t.session.EndSession(t.context)
	}
	t.logger.Info("[ MONGO ] - Transaction resources released")
}
