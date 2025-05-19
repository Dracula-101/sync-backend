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

// TransactionCallback defines the signature for functions that will be executed within a transaction
type TransactionCallback func(session DatabaseSession) error

type Transaction interface {
	Start() error
	Abort() error
	PerformTransaction(callback TransactionCallback) error
	IsDone() bool
	GetContext() context.Context
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

// NewTransaction creates a new MongoDB transaction with the specified timeout
func NewTransaction(logger utils.AppLogger, client *mongo.Client, database string, timeout time.Duration) Transaction {
	// We can't check utils.AppLogger directly for nil as it's an interface
	if client == nil {
		panic("mongo client is required for transaction")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &transaction{
		logger:     logger,
		database:   database,
		client:     client,
		context:    ctx,
		cancel:     cancel,
		hasTimeout: true,
	}
}

// NewDefaultTransaction creates a new MongoDB transaction with the default timeout
func NewDefaultTransaction(logger utils.AppLogger, client *mongo.Client, database string, timeout time.Duration) Transaction {
	return NewTransaction(logger, client, database, timeout)
}

func (t *transaction) Start() error {
	t.logger.Info("[ MONGO ] - Starting transaction")

	// Check if the context is already done (timed out)
	if t.IsDone() {
		t.logger.Error("[ MONGO ] - Cannot start transaction: context deadline exceeded")
		return fmt.Errorf("cannot start transaction: context deadline exceeded")
	}

	// Check if there's already an active session
	if t.session != nil {
		t.logger.Warn("[ MONGO ] - Session already exists, ending existing session before starting new one")
		t.session.EndSession(context.Background())
		t.session = nil
	}

	var err error
	t.session, err = t.client.StartSession()
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start session: %v", err)
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Verify connection to MongoDB with timeout
	err = t.session.Client().Ping(t.context, readpref.Primary())
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to ping primary: %v", err)
		t.session.EndSession(context.Background())
		t.session = nil
		return fmt.Errorf("failed to ping primary: %w", err)
	}

	// Configure transaction options
	txnOpts := options.Transaction()

	// Start the transaction with the options
	err = t.session.StartTransaction(txnOpts)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start transaction: %v", err)
		t.session.EndSession(context.Background())
		t.session = nil
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction started successfully")
	return nil
}

func (t *transaction) Abort() error {
	t.logger.Info("[ MONGO ] - Aborting transaction")
	defer t.cleanup() // Ensure the context is canceled and session is ended

	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot abort: no active session")
		return fmt.Errorf("cannot abort: no active session")
	}

	// Create a background context to use if the original context is done
	abortCtx := t.context
	if t.IsDone() {
		t.logger.Warn("[ MONGO ] - Transaction context is done, using background context for abort")
		var cancel context.CancelFunc
		abortCtx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	// Try to abort the transaction with the appropriate context
	err := t.session.AbortTransaction(abortCtx)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to abort transaction: %v", err)
		return fmt.Errorf("failed to abort transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction aborted successfully")
	return nil
}

// cleanup ensures that the context is canceled and the session is ended
// cleanup ensures that the context is canceled and the session is ended
// It creates a new context for ending the session if the original is already done
func (t *transaction) cleanup() {
	// Log that we're cleaning up resources
	t.logger.Info("[ MONGO ] - Cleaning up transaction resources")

	// Cancel the context if it exists
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	// End the session with a fresh context if it exists
	if t.session != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// We shouldn't ping during cleanup as it may not be necessary
		// and could cause delays if the server is having issues

		t.session.EndSession(cleanupCtx)
		t.session = nil
		t.logger.Info("[ MONGO ] - Session ended")
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

func (t *transaction) PerformTransaction(callback TransactionCallback) error {
	if t.session == nil {
		t.logger.Error("[ MONGO ] - Cannot perform transaction: no active session")
		return fmt.Errorf("cannot perform transaction: no active session")
	}

	if t.IsDone() {
		t.logger.Error("[ MONGO ] - Cannot perform transaction: context deadline exceeded")
		return fmt.Errorf("cannot perform transaction: context deadline exceeded")
	}

	defer t.cleanup() // Ensure resources are cleaned up after transaction

	result, err := t.session.WithTransaction(t.context, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		t.logger.Info("[ MONGO ] - Executing transaction callback")

		// Create an adapter that implements our DatabaseSession interface
		sessionAdapter := newSessionContextAdapter(sessionCtx, t.database)

		// Execute the callback with our abstraction layer
		err := callback(sessionAdapter)
		if err != nil {
			// No need to call t.Abort() here as WithTransaction handles this automatically
			t.logger.Error("[ MONGO ] - Transaction callback failed: %v", err)
			return nil, fmt.Errorf("transaction callback failed: %w", err)
		}

		t.logger.Info("[ MONGO ] - Transaction callback executed successfully")
		t.logger.Info("[ MONGO ] - Committing transaction")

		// Using the MongoDB transaction commit directly
		err = sessionCtx.CommitTransaction(sessionCtx)
		if err != nil {
			t.logger.Error("[ MONGO ] - Failed to commit transaction: %v", err)
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil, nil
	})

	if err != nil {
		t.logger.Error("[ MONGO ] - Transaction failed: %v", err)
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Use structured logging for complex objects
	if result != nil {
		t.logger.Info("[ MONGO ] - Transaction completed successfully")
	} else {
		t.logger.Info("[ MONGO ] - Transaction completed successfully with no result")
	}

	return nil
}
