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
type ClientHandle interface {
	Collection(name string) CollectionHandle
}

// CollectionHandle represents an abstracted MongoDB collection
type CollectionHandle interface {
	InsertOne(document interface{}) (interface{}, error)
	InsertMany(documents []interface{}) ([]interface{}, error)

	UpsertOne(filter interface{}, update interface{}) (int64, error)
	UpdateOne(filter interface{}, update interface{}) (int64, error)
	UpdateMany(filter interface{}, update interface{}) (int64, error)

	DeleteOne(filter interface{}) (int64, error)
	DeleteMany(filter interface{}) (int64, error)

	FindOne(filter interface{}) SingleResultHandle
	Find(filter interface{}) (MultipleResultHandle, error)

	FindOneAndUpdate(filter interface{}, update interface{}) SingleResultHandle
	FindOneAndDelete(filter interface{}) SingleResultHandle
	FindOneAndReplace(filter interface{}, replacement interface{}) SingleResultHandle
	CountDocuments(filter interface{}) (int64, error)

	Aggregate(pipeline interface{}) (MultipleResultHandle, error)
}

// SingleResultHandle represents a handle to a single document result
type SingleResultHandle interface {
	Decode(v interface{}) error
	Err() error
	IsNotFound() bool
}

// MultipleResultHandle represents a handle to multiple document results
type MultipleResultHandle interface {
	All(v interface{}) error
	Next() bool
	Decode(v interface{}) error
	Err() error
	Close() error
	TryNext(v interface{}) bool
	Remaining() int64
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
