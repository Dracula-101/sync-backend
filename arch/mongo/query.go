package mongo

import (
	"context"
	"fmt"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Query[T any] interface {
	Close()
	CheckIndexes(indexes []mongo.IndexModel) error
	CheckSearchIndexes(indexes []mongo.SearchIndexModel) error
	FindOne(filter bson.M, opts *options.FindOneOptions) (*T, error)
	FindAll(filter bson.M, opts *options.FindOptions) ([]*T, error)
	FindPaginated(filter bson.M, page int64, limit int64, opts *options.FindOptions) ([]*T, error)
	FindOneAndUpdate(filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (*T, error)
	FindOneAndReplace(filter bson.M, replacement *T, opts *options.FindOneAndReplaceOptions) (*T, error)
	FindOneAndDelete(filter bson.M, opts *options.FindOneAndDeleteOptions) (*T, error)
	InsertOne(doc *T) (*primitive.ObjectID, error)
	InsertAndRetrieveOne(doc *T) (*T, error)
	InsertMany(doc []*T) ([]primitive.ObjectID, error)
	InsertAndRetrieveMany(doc []*T) ([]*T, error)
	FilterOne(filter bson.M, opts *options.FindOneOptions) (*T, error)
	FilterOneAndUpdate(filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (*T, error)
	FilterMany(filter bson.M, opts *options.FindOptions) ([]*T, error)
	FilterPaginated(filter bson.M, page int64, limit int64, opts *options.FindOptions) ([]*T, error)
	FilterCount(filter bson.M) (int64, error)
	CountDocuments(filter bson.M, opts *options.CountOptions) (int64, error)
	UpdateOne(filter bson.M, update bson.M, opts *options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(filter bson.M, update bson.M, opts *options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(filter bson.M, opts *options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(filter bson.M, opts *options.DeleteOptions) (*mongo.DeleteResult, error)
}

type query[T any] struct {
	logger     utils.AppLogger
	collection *mongo.Collection
	context    context.Context
	cancel     context.CancelFunc
}

func newSingleQuery[T any](logger utils.AppLogger, collection *mongo.Collection, timeout time.Duration) Query[T] {
	context, cancel := context.WithTimeout(context.Background(), timeout)
	return &query[T]{
		logger:     logger,
		context:    context,
		cancel:     cancel,
		collection: collection,
	}
}

func newQuery[T any](logger utils.AppLogger, context context.Context, collection *mongo.Collection) Query[T] {
	return &query[T]{
		logger:     logger,
		context:    context,
		collection: collection,
	}
}

func (q *query[T]) Close() {
	if q.cancel != nil {
		q.cancel()
	}
}

func (q *query[T]) CheckIndexes(indexes []mongo.IndexModel) error {
	// get indexes and create if not exist and delete which are not in the list
	defer q.Close()
	q.logger.Info("[ MONGO ] - Checking indexes for collection: %s", q.collection.Name())
	existingIndexes, err := q.collection.Indexes().List(q.context)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error listing indexes: %v", err)
		return fmt.Errorf("error listing indexes: %w", err)
	}
	defer existingIndexes.Close(q.context)
	var existingIndexNames []string
	for existingIndexes.Next(q.context) {
		var index bson.M
		if err := existingIndexes.Decode(&index); err != nil {
			q.logger.Error("[ MONGO ] - Error decoding index: %v", err)
			return fmt.Errorf("error decoding index: %w", err)
		}
		if name, ok := index["name"].(string); ok {
			existingIndexNames = append(existingIndexNames, name)
		}
	}
	// Create indexes if not exist
	for _, index := range indexes {
		if !contains(existingIndexNames, index.Options.Name) {
			_, err := q.collection.Indexes().CreateOne(q.context, index)
			if err != nil {
				q.logger.Error("[ MONGO ] - Error creating index: %v", err)
				return fmt.Errorf("error creating index: %w", err)
			}
		}
	}
	// Delete indexes which are not in the list, but skip _id_ index
	for _, existingIndexName := range existingIndexNames {
		if existingIndexName == "_id_" {
			continue
		}
		if !contains(existingIndexNames, &existingIndexName) {
			_, err := q.collection.Indexes().DropOne(q.context, existingIndexName)
			if err != nil {
				q.logger.Error("[ MONGO ] - Error dropping index: %v", err)
				return fmt.Errorf("error dropping index: %w", err)
			}
		}
	}
	q.logger.Success("[ MONGO ] - Indexes checked for collection: %s", q.collection.Name())
	return nil
}

func (q *query[T]) CheckSearchIndexes(indexes []mongo.SearchIndexModel) error {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Checking search indexes for collection: %s", q.collection.Name())
	existingIndexes, err := q.collection.SearchIndexes().List(q.context, nil)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error listing search indexes: %v", err)
		return fmt.Errorf("error listing search indexes: %w", err)
	}
	defer existingIndexes.Close(q.context)
	var existingIndexNames []string
	for existingIndexes.Next(q.context) {
		var index bson.M
		if err := existingIndexes.Decode(&index); err != nil {
			q.logger.Error("[ MONGO ] - Error decoding search index: %v", err)
			return fmt.Errorf("error decoding search index: %w", err)
		}
		if name, ok := index["name"].(string); ok {
			existingIndexNames = append(existingIndexNames, name)
		}
	}
	// Create search indexes if not exist
	for _, index := range indexes {
		if !contains(existingIndexNames, index.Options.Name) {
			_, err := q.collection.SearchIndexes().CreateOne(q.context, index)
			if err != nil {
				q.logger.Error("[ MONGO ] - Error creating search index: %v", err)
				return fmt.Errorf("error creating search index: %w", err)
			}
		}
	}
	// Delete search indexes which are not in the list, but skip _id_ index
	for _, existingIndexName := range existingIndexNames {
		if !contains(existingIndexNames, &existingIndexName) {
			err := q.collection.SearchIndexes().DropOne(q.context, existingIndexName)
			if err != nil {
				q.logger.Error("[ MONGO ] - Error dropping search index: %v", err)
				return fmt.Errorf("error dropping search index: %w", err)
			}
		}
	}
	q.logger.Success("[ MONGO ] - Search indexes checked for collection: %s", q.collection.Name())
	return nil
}

func contains(existingIndexNames []string, string *string) bool {
	if string == nil {
		return false
	}
	for _, name := range existingIndexNames {
		if name == *string {
			return true
		}
	}
	return false
}

func (q *query[T]) FindOne(filter bson.M, opts *options.FindOneOptions) (*T, error) {
	defer q.Close()
	var doc T
	q.logger.Info("[ MONGO ] - Executing FindOne query with filter: %v", filter)
	err := q.collection.FindOne(q.context, filter, opts).Decode(&doc)
	if err != nil {
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FindOne query executed successfully")
	return &doc, nil
}

func (q *query[T]) FindAll(filter bson.M, opts *options.FindOptions) ([]*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FindAll query with filter: %v", filter)
	cursor, err := q.collection.Find(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindAll query: %v", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cursor.Close(q.context)

	var docs []*T

	for cursor.Next(q.context) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			q.logger.Error("[ MONGO ] - Error decoding result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		docs = append(docs, &result)
	}

	if err := cursor.Err(); err != nil {
		q.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	q.logger.Info("[ MONGO ] - FindAll query executed successfully, retrieved %d documents", len(docs))
	return docs, nil
}

func (q *query[T]) FindPaginated(filter bson.M, page int64, limit int64, opts *options.FindOptions) ([]*T, error) {
	defer q.Close()
	skip := (page - 1) * limit

	if opts == nil {
		opts = options.Find()
	}
	opts.SetSkip(skip)
	opts.SetLimit(int64(limit))
	q.logger.Info("[ MONGO ] - Executing FindPaginated query with filter: %v, page: %d, limit: %d", filter, page, limit)
	cursor, err := q.collection.Find(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindPaginated query: %v", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cursor.Close(q.context)

	var docs []*T

	for cursor.Next(q.context) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			q.logger.Error("[ MONGO ] - Error decoding result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		docs = append(docs, &result)
	}

	if err := cursor.Err(); err != nil {
		q.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	q.logger.Info("[ MONGO ] - FindPaginated query executed successfully, retrieved %d documents", len(docs))
	return docs, nil
}

func (q *query[T]) FindOneAndUpdate(filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FindOneAndUpdate query with filter: %v, update: %v", filter, update)
	var doc T
	err := q.collection.FindOneAndUpdate(q.context, filter, update, opts).Decode(&doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindOneAndUpdate query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FindOneAndUpdate query executed successfully")
	return &doc, nil
}

func (q *query[T]) FindOneAndReplace(filter bson.M, replacement *T, opts *options.FindOneAndReplaceOptions) (*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FindOneAndReplace query with filter: %v, replacement: %v", filter, replacement)
	var doc T
	err := q.collection.FindOneAndReplace(q.context, filter, replacement, opts).Decode(&doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindOneAndReplace query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FindOneAndReplace query executed successfully")
	return &doc, nil
}

func (q *query[T]) FindOneAndDelete(filter bson.M, opts *options.FindOneAndDeleteOptions) (*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FindOneAndDelete query with filter: %v", filter)
	var doc T
	err := q.collection.FindOneAndDelete(q.context, filter, opts).Decode(&doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindOneAndDelete query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FindOneAndDelete query executed successfully")
	return &doc, nil
}

func (q *query[T]) InsertOne(doc *T) (*primitive.ObjectID, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing InsertOne query with document")
	result, err := q.collection.InsertOne(q.context, doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing InsertOne query: %v", err)
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		q.logger.Error("[ MONGO ] - Error converting inserted ID to ObjectID: %v", result.InsertedID)
		return nil, fmt.Errorf("database query error for: %s", insertedID)
	}
	q.logger.Info("[ MONGO ] - InsertOne query executed successfully, inserted ID: %s", insertedID.Hex())
	return &insertedID, nil
}

func (q *query[T]) InsertAndRetrieveOne(doc *T) (*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing InsertAndRetrieveOne query")
	result, err := q.collection.InsertOne(q.context, doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing InsertAndRetrieveOne query: %v", err)
		return nil, err
	}

	filter := bson.M{"_id": result.InsertedID}
	q.logger.Info("[ MONGO ] - Executing FindOne query with filter: %v", filter)
	retrived, err := q.FindOne(filter, nil)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindOne query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - InsertAndRetrieveOne query executed successfully")
	return retrived, nil
}

func (q *query[T]) InsertMany(docs []*T) ([]primitive.ObjectID, error) {
	defer q.Close()
	var iDocs []any
	for _, doc := range docs {
		iDocs = append(iDocs, doc)
	}
	q.logger.Info("[ MONGO ] - Executing InsertMany query")
	result, err := q.collection.InsertMany(q.context, iDocs)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing InsertMany query: %v", err)
		return nil, err
	}

	var insertedIDs []primitive.ObjectID

	for _, v := range result.InsertedIDs {
		insertedID, ok := v.(primitive.ObjectID)
		if !ok {
			q.logger.Error("[ MONGO ] - Error converting inserted ID to ObjectID: %v", v)
			return nil, fmt.Errorf("database query error for: %s", insertedID)
		}
		insertedIDs = append(insertedIDs, insertedID)
	}
	q.logger.Info("[ MONGO ] - InsertMany query executed successfully, inserted IDs: %v", insertedIDs)
	return insertedIDs, nil
}

func (q *query[T]) InsertAndRetrieveMany(docs []*T) ([]*T, error) {
	defer q.Close()
	var iDocs []any
	for _, doc := range docs {
		iDocs = append(iDocs, doc)
	}
	q.logger.Info("[ MONGO ] - Executing InsertAndRetrieveMany query")
	result, err := q.collection.InsertMany(q.context, iDocs)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing InsertAndRetrieveMany query: %v", err)
		return nil, err
	}

	filter := bson.M{"_id": bson.M{"$in": result.InsertedIDs}}
	q.logger.Info("[ MONGO ] - Executing FindAll query with filter: %v", filter)
	retrieved, err := q.FindAll(filter, nil)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FindAll query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - InsertAndRetrieveMany query executed successfully")
	return retrieved, nil
}

func (q *query[T]) FilterOne(filter bson.M, opts *options.FindOneOptions) (*T, error) {
	defer q.Close()
	var doc T
	q.logger.Info("[ MONGO ] - Executing FilterOne query with filter: %v", filter)
	err := q.collection.FindOne(q.context, filter, opts).Decode(&doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FilterOne query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FilterOne query executed successfully")
	return &doc, nil
}

func (q *query[T]) FilterOneAndUpdate(filter bson.M, update bson.M, opts *options.FindOneAndUpdateOptions) (*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FilterOneAndUpdate query with filter: %v, update: %v", filter, update)
	var doc T
	err := q.collection.FindOneAndUpdate(q.context, filter, update, opts).Decode(&doc)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FilterOneAndUpdate query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - FilterOneAndUpdate query executed successfully")
	return &doc, nil
}

func (q *query[T]) FilterMany(filter bson.M, opts *options.FindOptions) ([]*T, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FilterMany query with filter: %v", filter)
	cursor, err := q.collection.Find(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FilterMany query: %v", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cursor.Close(q.context)

	var docs []*T

	for cursor.Next(q.context) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			q.logger.Error("[ MONGO ] - Error decoding result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		docs = append(docs, &result)
	}

	if err := cursor.Err(); err != nil {
		q.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	q.logger.Info("[ MONGO ] - FilterMany query executed successfully, retrieved %d documents", len(docs))
	return docs, nil
}

func (q *query[T]) FilterPaginated(filter bson.M, page int64, limit int64, opts *options.FindOptions) ([]*T, error) {
	defer q.Close()
	skip := (page - 1) * limit

	if opts == nil {
		opts = options.Find()
	}
	opts.SetSkip(skip)
	opts.SetLimit(int64(limit))
	q.logger.Info("[ MONGO ] - Executing FilterPaginated query with filter: %v, page: %d, limit: %d", filter, page, limit)
	cursor, err := q.collection.Find(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FilterPaginated query: %v", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cursor.Close(q.context)

	var docs []*T

	for cursor.Next(q.context) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			q.logger.Error("[ MONGO ] - Error decoding result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		docs = append(docs, &result)
	}

	if err := cursor.Err(); err != nil {
		q.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	q.logger.Info("[ MONGO ] - FilterPaginated query executed successfully, retrieved %d documents", len(docs))
	return docs, nil
}

func (q *query[T]) FilterCount(filter bson.M) (int64, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing FilterCount query with filter: %v", filter)
	count, err := q.collection.CountDocuments(q.context, filter)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing FilterCount query: %v", err)
		return 0, fmt.Errorf("error executing query: %w", err)
	}
	q.logger.Info("[ MONGO ] - FilterCount query executed successfully, count: %d", count)
	return count, nil
}

func (q *query[T]) CountDocuments(filter bson.M, opts *options.CountOptions) (int64, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing CountDocuments query with filter: %v", filter)
	count, err := q.collection.CountDocuments(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing CountDocuments query: %v", err)
		return 0, fmt.Errorf("error executing query: %w", err)
	}
	q.logger.Info("[ MONGO ] - CountDocuments query executed successfully, count: %d", count)
	return count, nil
}

/*
 * Example -> update := bson.M{"$set": bson.M{"field": "newValue"}}
 */
func (q *query[T]) UpdateOne(filter bson.M, update bson.M, opts *options.UpdateOptions) (*mongo.UpdateResult, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing UpdateOne query with filter: %v, update: %v", filter, update)
	result, err := q.collection.UpdateOne(q.context, filter, update, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing UpdateOne query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - UpdateOne query executed successfully, modified count: %d", result.ModifiedCount)
	return result, nil
}

/*
 * Example -> update := bson.M{"$set": bson.M{"field": "newValue"}}
 */
func (q *query[T]) UpdateMany(filter bson.M, update bson.M, opts *options.UpdateOptions) (*mongo.UpdateResult, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing UpdateMany query with filter: %v, update: %v", filter, update)
	result, err := q.collection.UpdateMany(q.context, filter, update, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing UpdateMany query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - UpdateMany query executed successfully, modified count: %d", result.ModifiedCount)
	return result, nil
}

func (q *query[T]) DeleteOne(filter bson.M, opts *options.DeleteOptions) (*mongo.DeleteResult, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing DeleteOne query with filter: %v", filter)
	result, err := q.collection.DeleteOne(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing DeleteOne query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - DeleteOne query executed successfully, deleted count: %d", result.DeletedCount)
	return result, nil
}

func (q *query[T]) DeleteMany(filter bson.M, opts *options.DeleteOptions) (*mongo.DeleteResult, error) {
	defer q.Close()
	q.logger.Info("[ MONGO ] - Executing DeleteMany query with filter: %v", filter)
	result, err := q.collection.DeleteMany(q.context, filter, opts)
	if err != nil {
		q.logger.Error("[ MONGO ] - Error executing DeleteMany query: %v", err)
		return nil, err
	}
	q.logger.Info("[ MONGO ] - DeleteMany query executed successfully, deleted count: %d", result.DeletedCount)
	return result, nil
}
