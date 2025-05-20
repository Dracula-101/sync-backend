package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionSession represents an abstracted MongoDB session for transactions
// It provides access to collections without requiring MongoDB-specific dependencies
type TransactionSession interface {
	// Collection returns a handle to a collection in the default database
	Collection(name string) CollectionHandle
	// Client returns a handle to the database client
	Client() ClientHandle
}

// ClientHandle represents an abstracted MongoDB client
// It allows access to collections through the abstraction layer
type ClientHandle interface {
	// Collection returns a handle to a collection in the default database
	Collection(name string) CollectionHandle
}

// CollectionHandle represents an abstracted MongoDB collection
// It provides all common operations for working with collections without MongoDB dependencies
type CollectionHandle interface {
	// InsertOne inserts a single document and returns its ID
	InsertOne(document interface{}) (interface{}, error)
	// InsertMany inserts multiple documents and returns their IDs
	InsertMany(documents []interface{}) ([]interface{}, error)

	// UpdateOne updates a single document and returns the number of modified documents
	UpsertOne(filter interface{}, update interface{}) (int64, error)
	UpdateOne(filter interface{}, update interface{}) (int64, error)
	// UpdateMany updates multiple documents and returns the number of modified documents
	UpdateMany(filter interface{}, update interface{}) (int64, error)
	// DeleteOne deletes a single document and returns the number of deleted documents
	DeleteOne(filter interface{}) (int64, error)
	// DeleteMany deletes multiple documents and returns the number of deleted documents
	DeleteMany(filter interface{}) (int64, error)
	// FindOne finds a single document matching the filter
	FindOne(filter interface{}) SingleResultHandle
	// Find finds all documents matching the filter
	Find(filter interface{}) (MultipleResultHandle, error)
	// FindOneAndUpdate finds a document and updates it, returning the updated document
	FindOneAndUpdate(filter interface{}, update interface{}) SingleResultHandle
	// FindOneAndDelete finds a document and deletes it, returning the deleted document
	FindOneAndDelete(filter interface{}) SingleResultHandle
	// FindOneAndReplace finds a document and replaces it, returning the new document
	FindOneAndReplace(filter interface{}, replacement interface{}) SingleResultHandle
	// CountDocuments counts the number of documents matching the filter
	CountDocuments(filter interface{}) (int64, error)
	// Aggregate performs an aggregation pipeline operation
	Aggregate(pipeline interface{}) (MultipleResultHandle, error)
}

// SingleResultHandle represents a handle to a single document result
// It provides methods for error handling and decoding the document
type SingleResultHandle interface {
	// Decode decodes the result into the provided value
	Decode(v interface{}) error
	// Err returns any error encountered during the operation
	Err() error
	// IsNotFound returns true if no document was found
	IsNotFound() bool
}

// MultipleResultHandle represents a handle to multiple document results
// It provides methods for iterating through results and error handling
type MultipleResultHandle interface {
	// All decodes all results into the provided slice
	All(v interface{}) error
	// Next advances the cursor to the next document and returns true if one exists
	Next() bool
	// Decode decodes the current document into the provided value
	Decode(v interface{}) error
	// Err returns any error encountered during iteration
	Err() error
	// Close releases resources associated with the cursor
	Close() error
	// TryNext attempts to advance the cursor and decode in one operation
	TryNext(v interface{}) bool
	// Remaining returns the number of documents left in the current batch
	Remaining() int64
	// ID returns the cursor ID
	ID() interface{}
}

type sessionContextAdapter struct {
	ctx      mongo.SessionContext
	database string
}

func newSessionContextAdapter(ctx mongo.SessionContext, database string) TransactionSession {
	return &sessionContextAdapter{
		ctx:      ctx,
		database: database,
	}
}

func (s *sessionContextAdapter) Collection(name string) CollectionHandle {
	return &collectionAdapter{
		coll: s.ctx.Client().Database(s.database).Collection(name),
		ctx:  s.ctx,
	}
}

func (s *sessionContextAdapter) Client() ClientHandle {
	return &clientAdapter{client: s.ctx.Client(), ctx: s.ctx, database: s.database}
}

type clientAdapter struct {
	client   *mongo.Client
	ctx      mongo.SessionContext
	database string
}

func (c *clientAdapter) Collection(name string) CollectionHandle {
	return &collectionAdapter{
		coll: c.client.Database(c.database).Collection(name),
		ctx:  c.ctx,
	}
}

type collectionAdapter struct {
	coll *mongo.Collection
	ctx  mongo.SessionContext
}

func (c *collectionAdapter) InsertOne(document interface{}) (interface{}, error) {
	result, err := c.coll.InsertOne(c.ctx, document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (c *collectionAdapter) InsertMany(documents []interface{}) ([]interface{}, error) {
	result, err := c.coll.InsertMany(c.ctx, documents)
	if err != nil {
		return nil, err
	}
	return result.InsertedIDs, nil
}

func (c *collectionAdapter) UpsertOne(filter interface{}, update interface{}) (int64, error) {
	opts := options.Update().SetUpsert(true)
	result, err := c.coll.UpdateOne(c.ctx, filter, update, opts)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (c *collectionAdapter) UpdateOne(filter interface{}, update interface{}) (int64, error) {
	result, err := c.coll.UpdateOne(c.ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (c *collectionAdapter) UpdateMany(filter interface{}, update interface{}) (int64, error) {
	result, err := c.coll.UpdateMany(c.ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (c *collectionAdapter) DeleteOne(filter interface{}) (int64, error) {
	result, err := c.coll.DeleteOne(c.ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (c *collectionAdapter) DeleteMany(filter interface{}) (int64, error) {
	result, err := c.coll.DeleteMany(c.ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (c *collectionAdapter) FindOne(filter interface{}) SingleResultHandle {
	result := c.coll.FindOne(c.ctx, filter)
	return &singleResultAdapter{result: result}
}

func (c *collectionAdapter) Find(filter interface{}) (MultipleResultHandle, error) {
	cursor, err := c.coll.Find(c.ctx, filter)
	if err != nil {
		return nil, err
	}
	return &multipleResultAdapter{cursor: cursor, ctx: c.ctx}, nil
}

func (c *collectionAdapter) FindOneAndUpdate(filter interface{}, update interface{}) SingleResultHandle {
	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)
	result := c.coll.FindOneAndUpdate(c.ctx, filter, update, opts)
	return &singleResultAdapter{result: result}
}

func (c *collectionAdapter) FindOneAndDelete(filter interface{}) SingleResultHandle {
	result := c.coll.FindOneAndDelete(c.ctx, filter)
	return &singleResultAdapter{result: result}
}

func (c *collectionAdapter) FindOneAndReplace(filter interface{}, replacement interface{}) SingleResultHandle {
	opts := options.FindOneAndReplace()
	opts.SetReturnDocument(options.After)
	result := c.coll.FindOneAndReplace(c.ctx, filter, replacement, opts)
	return &singleResultAdapter{result: result}
}

func (c *collectionAdapter) CountDocuments(filter interface{}) (int64, error) {
	count, err := c.coll.CountDocuments(c.ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *collectionAdapter) Aggregate(pipeline interface{}) (MultipleResultHandle, error) {
	cursor, err := c.coll.Aggregate(c.ctx, pipeline)
	if err != nil {
		return nil, err
	}
	return &multipleResultAdapter{cursor: cursor}, nil
}

type multipleResultAdapter struct {
	cursor *mongo.Cursor
	ctx    context.Context
}

func (m *multipleResultAdapter) All(v interface{}) error {
	return m.cursor.All(m.ctx, v)
}

func (m *multipleResultAdapter) Next() bool {
	return m.cursor.Next(m.ctx)
}

func (m *multipleResultAdapter) TryNext(v interface{}) bool {
	if m.cursor.TryNext(m.ctx) {
		if v != nil {
			if err := m.cursor.Decode(v); err != nil {
				return false
			}
		}
		return true
	}
	return false
}

func (m *multipleResultAdapter) Decode(v interface{}) error {
	return m.cursor.Decode(v)
}

func (m *multipleResultAdapter) Err() error {
	return m.cursor.Err()
}

func (m *multipleResultAdapter) Close() error {
	return m.cursor.Close(m.ctx)
}

func (m *multipleResultAdapter) Remaining() int64 {
	return int64(m.cursor.RemainingBatchLength())
}

func (m *multipleResultAdapter) ID() interface{} {
	return m.cursor.ID()
}

type singleResultAdapter struct {
	result *mongo.SingleResult
}

func (s *singleResultAdapter) Decode(v interface{}) error {
	return s.result.Decode(v)
}

func (s *singleResultAdapter) Err() error {
	return s.result.Err()
}

func (s *singleResultAdapter) IsNotFound() bool {
	return s.result.Err() == mongo.ErrNoDocuments
}
