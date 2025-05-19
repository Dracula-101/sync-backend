package mongo

import (
	"context"
	"fmt"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionCallback func(sessionCtx TransactionSession) error

type Transaction interface {
	Start() error
	Abort() error
	Commit() error
	IsDone() bool

	/* INSERT OPERATIONS */
	InsertOne(collectionName string, document interface{}) (interface{}, error)
	InsertMany(collectionName string, documents []interface{}) ([]interface{}, error)

	/* FINDING OPERATIONS */
	FindOne(collectionName string, filter interface{}, result interface{}) error
	FindMany(collectionName string, filter interface{}, result interface{}) error

	/* UDPATE OPERATIONS */
	UpdateOne(collectionName string, filter interface{}, update interface{}) (int64, error)
	UpdateMany(collectionName string, filter interface{}, update interface{}) (int64, error)

	/* DELETE OPERATIONS */
	DeleteOne(collectionName string, filter interface{}) (int64, error)
	DeleteMany(collectionName string, filter interface{}) (int64, error)

	/* FIND AND UPDATE OPERATIONS */
	FindOneAndUpdate(collectionName string, filter interface{}, update interface{}, result interface{}) error
	FindOneAndDelete(collectionName string, filter interface{}, result interface{}) error

	/* COUNT */
	CountDocuments(collectionName string, filter interface{}) (int64, error)

	/* SINGLE TRANSACTION */
	PerformSingleTransaction(callback TransactionCallback) error
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
		database:   database,
		client:     client,
		context:    ctx,
		cancel:     cancel,
		hasTimeout: true,
	}
}

func (t *transaction) Start() error {
	if t.session != nil {
		t.logger.Info("[ MONGO ] - Transaction already started")
		return nil
	}

	var err error
	t.session, err = t.client.StartSession()
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start session: %v", err)
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = t.session.StartTransaction()
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to start transaction: %v", err)
		t.session.EndSession(t.context)
		t.session = nil
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	t.logger.Info("[ MONGO ] - Transaction started successfully")
	return nil
}

func (t *transaction) Abort() error {
	if t.session == nil {
		t.logger.Info("[ MONGO ] - No active transaction to abort (already aborted or committed)")
		return nil
	}

	err := t.session.AbortTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to abort transaction: %v", err)
		return fmt.Errorf("failed to abort transaction: %w", err)
	}

	t.session.EndSession(t.context)
	t.session = nil
	t.logger.Info("[ MONGO ] - Transaction aborted successfully")
	return nil
}

func (t *transaction) Commit() error {
	if t.session == nil {
		t.logger.Error("[ MONGO ] - No active transaction to commit")
		return fmt.Errorf("no active transaction to commit")
	}

	err := t.session.CommitTransaction(t.context)
	if err != nil {
		t.logger.Error("[ MONGO ] - Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.session.EndSession(t.context)
	t.session = nil
	t.logger.Info("[ MONGO ] - Transaction committed successfully")
	return nil
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

func (t *transaction) PerformSingleTransaction(callback TransactionCallback) error {
	sessOpts := options.Session()
	newSession, err := t.client.StartSession(sessOpts)
	if err != nil {
		return err
	}
	defer newSession.EndSession(t.context)

	_, err = newSession.WithTransaction(t.context, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		t.logger.Info("[ MONGO ] - Executing transaction callback")
		sessionAdapter := newSessionContextAdapter(sessionCtx, t.database)
		if sessionAdapter == nil {
			t.logger.Error("[ MONGO ] - Failed to create session context adapter")
			return nil, fmt.Errorf("failed to create session context adapter")
		}

		err = callback(sessionAdapter)
		if err != nil {
			t.logger.Error("[ MONGO ] - Transaction callback failed: %v", err)
			return nil, err
		}

		t.logger.Info("[ MONGO ] - Committing transaction")
		return nil, nil
	})

	if err != nil {
		t.logger.Error("[ MONGO ] - Transaction failed: %v", err)
	}

	t.logger.Info("[ MONGO ] - Transaction completed successfully")
	return err
}

/* INSERT OPERATIONS */
func (t *transaction) InsertOne(collectionName string, document interface{}) (interface{}, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	results, err := collection.InsertOne(t.context, document)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to insert document: %v", err)
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document inserted successfully: %v", results.InsertedID)
	return results.InsertedID, nil
}

func (t *transaction) InsertMany(collectionName string, documents []interface{}) ([]interface{}, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	results, err := collection.InsertMany(t.context, documents)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to insert documents: %v", err)
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Documents inserted successfully: %v", results.InsertedIDs)
	return results.InsertedIDs, nil
}

func (t *transaction) FindOne(collectionName string, filter interface{}, result interface{}) error {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	err := collection.FindOne(t.context, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t.logger.Info("[ MONGO ] - [ TRANSACTION ] No document found for filter: %v", filter)
			return nil
		}
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to find document: %v", err)
		return fmt.Errorf("failed to find document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document found successfully: %v", result)
	return nil
}

func (t *transaction) FindMany(collectionName string, filter interface{}, result interface{}) error {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	cursor, err := collection.Find(t.context, filter)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to find documents: %v", err)
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(t.context)
	err = cursor.All(t.context, result)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to decode documents: %v", err)
		return fmt.Errorf("failed to decode documents: %w", err)
	}

	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Documents found successfully: %v", result)
	return nil
}

func (t *transaction) UpdateOne(collectionName string, filter interface{}, update interface{}) (int64, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	result, err := collection.UpdateOne(t.context, filter, update)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to update document: %v", err)
		return 0, fmt.Errorf("failed to update document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document updated successfully: %v", result.ModifiedCount)
	return result.ModifiedCount, nil
}

func (t *transaction) UpdateMany(collectionName string, filter interface{}, update interface{}) (int64, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	result, err := collection.UpdateMany(t.context, filter, update)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to update documents: %v", err)
		return 0, fmt.Errorf("failed to update documents: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Documents updated successfully: %v", result.ModifiedCount)
	return result.ModifiedCount, nil
}

func (t *transaction) DeleteOne(collectionName string, filter interface{}) (int64, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	result, err := collection.DeleteOne(t.context, filter)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to delete document: %v", err)
		return 0, fmt.Errorf("failed to delete document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document deleted successfully: %v", result.DeletedCount)
	return result.DeletedCount, nil
}

func (t *transaction) DeleteMany(collectionName string, filter interface{}) (int64, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	result, err := collection.DeleteMany(t.context, filter)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to delete documents: %v", err)
		return 0, fmt.Errorf("failed to delete documents: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Documents deleted successfully: %v", result.DeletedCount)
	return result.DeletedCount, nil
}

func (t *transaction) FindOneAndUpdate(collectionName string, filter interface{}, update interface{}, result interface{}) error {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	err := collection.FindOneAndUpdate(t.context, filter, update).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t.logger.Info("[ MONGO ] - [ TRANSACTION ] No document found for filter: %v", filter)
			return nil
		}
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to find and update document: %v", err)
		return fmt.Errorf("failed to find and update document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document found and updated successfully: %v", result)
	return nil
}

func (t *transaction) FindOneAndDelete(collectionName string, filter interface{}, result interface{}) error {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	err := collection.FindOneAndDelete(t.context, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t.logger.Info("[ MONGO ] - [ TRANSACTION ] No document found for filter: %v", filter)
			return nil
		}
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to find and delete document: %v", err)
		return fmt.Errorf("failed to find and delete document: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document found and deleted successfully: %v", result)
	return nil
}

func (t *transaction) CountDocuments(collectionName string, filter interface{}) (int64, error) {
	collection := t.session.Client().Database(t.database).Collection(collectionName)
	count, err := collection.CountDocuments(t.context, filter)
	if err != nil {
		t.logger.Error("[ MONGO ] - [ TRANSACTION ] Failed to count documents: %v", err)
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	t.logger.Info("[ MONGO ] - [ TRANSACTION ] Document count: %v", count)
	return count, nil
}
