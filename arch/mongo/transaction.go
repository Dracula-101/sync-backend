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
	Commit() error
	Abort() error
	GetContext() context.Context
	GetCollection(name string) *mongo.Collection
	IsDone() bool
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

const DefaultShortTransactionTimeout = 30 * time.Second
const DefaultTransactionTimeout = 1 * time.Minute
const DefaultLongTransactionTimeout = 5 * time.Minute

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

func (t *transaction) Start() error {
	t.logger.Info("[ MONGO ] - Starting transaction")

	// Check if the context is already done (timed out)
	if t.IsDone() {
		t.logger.Error("[ MONGO ] - Cannot start transaction: context deadline exceeded")
		return fmt.Errorf("cannot start transaction: context deadline exceeded")
	}

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

func (t *transaction) Commit() error {
	t.logger.Info("[ MONGO ] - Committing transaction")
	defer t.cleanup() // Ensure the context is canceled and session is ended

	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot commit: no active session")
		return fmt.Errorf("cannot commit: no active session")
	}

	// Check if the context is already done (timed out)
	if t.IsDone() {
		t.logger.Error("[ MONGO ] - Cannot commit transaction: context deadline exceeded")
		return fmt.Errorf("cannot commit transaction: context deadline exceeded")
	}

	err := t.session.CommitTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction committed successfully")
	return nil
}

func (t *transaction) Abort() error {
	t.logger.Info("[ MONGO ] - Aborting transaction")
	defer t.cleanup() // Ensure the context is canceled and session is ended

	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot abort: no active session")
		return fmt.Errorf("cannot abort: no active session")
	}

	// Even if the context is done, we should try to abort the transaction
	err := t.session.AbortTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to abort transaction: %v", err)
		return fmt.Errorf("failed to abort transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction aborted successfully")
	return nil
}

// cleanup ensures that the context is canceled and the session is ended
func (t *transaction) cleanup() {
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	if t.session != nil {
		t.session.EndSession(context.Background()) // Use a new context in case the original is already done
		t.session = nil
	}
}

func (t *transaction) GetContext() context.Context {
	return t.context
}

// IsDone checks if the transaction's context is already done (timed out or canceled)
func (t *transaction) IsDone() bool {
	select {
	case <-t.context.Done():
		return true
	default:
		return false
	}
}

func (t *transaction) GetCollection(name string) *mongo.Collection {
	return t.client.Database(t.database).Collection(name)
}
