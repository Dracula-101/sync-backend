package mongo

import (
	"context"
	"fmt"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StageType represents the type of aggregation stage
type StageType string

// MongoDB aggregation stage types
const (
	StageSearch          StageType = "$search"
	StageMatch           StageType = "$match"
	StageGroup           StageType = "$group"
	StageSort            StageType = "$sort"
	StageProject         StageType = "$project"
	StageAddFields       StageType = "$addFields"
	StageUnwind          StageType = "$unwind"
	StageLookup          StageType = "$lookup"
	StageLimit           StageType = "$limit"
	StageSkip            StageType = "$skip"
	StageCount           StageType = "$count"
	StageFacet           StageType = "$facet"
	StageGraphLookup     StageType = "$graphLookup"
	StageReplaceWith     StageType = "$replaceWith"
	StageReplaceRoot     StageType = "$replaceRoot"
	StageUnset           StageType = "$unset"
	StageRedact          StageType = "$redact"
	StageSample          StageType = "$sample"
	StageOut             StageType = "$out"
	StageSetWindowFields StageType = "$setWindowFields"
	StageGeoNear         StageType = "$geoNear"
	StageBucketAuto      StageType = "$bucketAuto"
	StageBucket          StageType = "$bucket"
)

// Aggregator provides a chainable API for MongoDB aggregation operations
type Aggregator[T any, R any] interface {
	// Configuration
	AllowDiskUse(allow bool) Aggregator[T, R]

	// Pipeline construction methods
	Search(index string, query interface{}) Aggregator[T, R]
	Match(filter interface{}) Aggregator[T, R]
	Project(projection interface{}) Aggregator[T, R]
	Group(groupBy interface{}) Aggregator[T, R]
	Sort(sortBy interface{}) Aggregator[T, R]
	Skip(n int64) Aggregator[T, R]
	Limit(n int64) Aggregator[T, R]
	AddFields(fields interface{}) Aggregator[T, R]
	Unwind(path string, options ...interface{}) Aggregator[T, R]
	Count(fieldName string) Aggregator[T, R]
	Facet(facets map[string][]bson.D) Aggregator[T, R]
	ReplaceRoot(newRoot interface{}) Aggregator[T, R]
	ReplaceWith(expression interface{}) Aggregator[T, R]
	Sample(size int64) Aggregator[T, R]
	Unset(fields ...string) Aggregator[T, R]
	Redact(expression interface{}) Aggregator[T, R]

	// Lookup related operations
	Lookup(from, localField, foreignField, as string) Aggregator[T, R]
	GraphLookup(from, startWith, connectFrom, connectTo, as string, options ...interface{}) Aggregator[T, R]

	// Geo operations
	GeoNear(coordinates [2]float64, options interface{}) Aggregator[T, R]

	// Pipeline execution methods
	Exec() ([]*R, error)
	ExecOne() (*R, error)
	ExecPaginated(page, limit int64) ([]*R, error)
	ExecCount() (int64, error)
	ExecRaw() (interface{}, error)

	// Get pipeline
	GetPipeline() []bson.D

	// Release resources
	Close()
}

// aggregator implements the Aggregator interface
type aggregator[T any, R any] struct {
	logger       utils.AppLogger
	collection   *mongo.Collection
	context      context.Context
	cancel       context.CancelFunc
	pipeline     []bson.D
	stageMap     map[StageType]int // tracks which stages exist and their positions
	allowDiskUse bool
}

// newSingleAggregator creates a new aggregator with a timeout
func newSingleAggregator[T any, R any](
	logger utils.AppLogger,
	collection *mongo.Collection,
	timeout time.Duration,
) Aggregator[T, R] {
	context, cancel := context.WithTimeout(context.Background(), timeout)
	return &aggregator[T, R]{
		logger:       logger,
		context:      context,
		cancel:       cancel,
		collection:   collection,
		pipeline:     make([]bson.D, 0),
		stageMap:     make(map[StageType]int),
		allowDiskUse: true,
	}
}

// newAggregator creates a new aggregator with an existing context
func newAggregator[T any, R any](
	logger utils.AppLogger,
	context context.Context,
	collection *mongo.Collection,
) Aggregator[T, R] {
	return &aggregator[T, R]{
		logger:       logger,
		context:      context,
		collection:   collection,
		pipeline:     make([]bson.D, 0),
		stageMap:     make(map[StageType]int),
		allowDiskUse: true,
	}
}

// Close releases resources held by the aggregator
func (a *aggregator[T, R]) Close() {
	if a.cancel != nil {
		a.cancel()
	}
}

// GetPipeline returns the current aggregation pipeline
func (a *aggregator[T, R]) GetPipeline() []bson.D {
	return a.pipeline
}

// addStage adds a stage to the pipeline, replacing existing stage of the same type if overwriteExisting is true
func (a *aggregator[T, R]) addStage(stageType StageType, operator string, value interface{}, overwriteExisting bool) {
	if pos, exists := a.stageMap[stageType]; exists && overwriteExisting {
		// Replace existing stage
		a.pipeline[pos] = bson.D{{Key: operator, Value: value}}
		return
	}

	// Add new stage to pipeline
	a.pipeline = append(a.pipeline, bson.D{{Key: operator, Value: value}})
	a.stageMap[stageType] = len(a.pipeline) - 1
}

// AllowDiskUse sets whether the aggregation can use disk for temporary storage
func (a *aggregator[T, R]) AllowDiskUse(allow bool) Aggregator[T, R] {
	a.allowDiskUse = allow
	return a
}

// Search adds a $search stage for MongoDB Atlas Search
func (a *aggregator[T, R]) Search(index string, query interface{}) Aggregator[T, R] {
	searchDoc := bson.M{"index": index}

	// If query is a map, merge it with the index
	if queryMap, ok := query.(bson.M); ok {
		for k, v := range queryMap {
			searchDoc[k] = v
		}
	} else {
		// If it's not a map, assume it's a full search definition
		searchDoc = bson.M{"index": index, "compound": query}
	}

	a.addStage(StageSearch, "$search", searchDoc, true)
	return a
}

// Match adds a $match stage
func (a *aggregator[T, R]) Match(filter interface{}) Aggregator[T, R] {
	a.addStage(StageMatch, "$match", filter, false)
	return a
}

// Project adds a $project stage
func (a *aggregator[T, R]) Project(projection interface{}) Aggregator[T, R] {
	a.addStage(StageProject, "$project", projection, false)
	return a
}

// Group adds a $group stage
func (a *aggregator[T, R]) Group(groupBy interface{}) Aggregator[T, R] {
	a.addStage(StageGroup, "$group", groupBy, false)
	return a
}

// Sort adds a $sort stage
func (a *aggregator[T, R]) Sort(sortBy interface{}) Aggregator[T, R] {
	a.addStage(StageSort, "$sort", sortBy, false)
	return a
}

// Skip adds a $skip stage
func (a *aggregator[T, R]) Skip(n int64) Aggregator[T, R] {
	a.addStage(StageSkip, "$skip", n, true)
	return a
}

// Limit adds a $limit stage
func (a *aggregator[T, R]) Limit(n int64) Aggregator[T, R] {
	a.addStage(StageLimit, "$limit", n, true)
	return a
}

// AddFields adds a $addFields stage
func (a *aggregator[T, R]) AddFields(fields interface{}) Aggregator[T, R] {
	a.addStage(StageAddFields, "$addFields", fields, false)
	return a
}

// Unwind adds a $unwind stage
func (a *aggregator[T, R]) Unwind(path string, options ...interface{}) Aggregator[T, R] {
	if len(options) > 0 {
		// Advanced unwind with options
		unwindDoc := bson.M{"path": "$" + path}

		// Handle options
		for _, opt := range options {
			if optMap, ok := opt.(map[string]interface{}); ok {
				for k, v := range optMap {
					unwindDoc[k] = v
				}
			}
		}

		a.addStage(StageUnwind, "$unwind", unwindDoc, false)
	} else {
		// Simple unwind
		a.addStage(StageUnwind, "$unwind", "$"+path, false)
	}
	return a
}

// Count adds a $count stage
func (a *aggregator[T, R]) Count(fieldName string) Aggregator[T, R] {
	a.addStage(StageCount, "$count", fieldName, true)
	return a
}

// Facet adds a $facet stage
func (a *aggregator[T, R]) Facet(facets map[string][]bson.D) Aggregator[T, R] {
	facetDoc := bson.D{}
	for name, subPipeline := range facets {
		facetDoc = append(facetDoc, bson.E{Key: name, Value: subPipeline})
	}
	a.addStage(StageFacet, "$facet", facetDoc, true)
	return a
}

// Lookup adds a $lookup stage for joining with another collection
func (a *aggregator[T, R]) Lookup(from, localField, foreignField, as string) Aggregator[T, R] {
	lookupDoc := bson.D{
		{Key: "from", Value: from},
		{Key: "localField", Value: localField},
		{Key: "foreignField", Value: foreignField},
		{Key: "as", Value: as},
	}
	a.addStage(StageLookup, "$lookup", lookupDoc, false)
	return a
}

// GraphLookup adds a $graphLookup stage for recursive lookups
func (a *aggregator[T, R]) GraphLookup(from, startWith, connectFrom, connectTo, as string, options ...interface{}) Aggregator[T, R] {
	graphLookupDoc := bson.D{
		{Key: "from", Value: from},
		{Key: "startWith", Value: "$" + startWith},
		{Key: "connectFromField", Value: connectFrom},
		{Key: "connectToField", Value: connectTo},
		{Key: "as", Value: as},
	}

	// Process additional options
	if len(options) > 0 {
		if optMap, ok := options[0].(map[string]interface{}); ok {
			for k, v := range optMap {
				graphLookupDoc = append(graphLookupDoc, bson.E{Key: k, Value: v})
			}
		}
	}

	a.addStage(StageGraphLookup, "$graphLookup", graphLookupDoc, false)
	return a
}

// ReplaceRoot adds a $replaceRoot stage
func (a *aggregator[T, R]) ReplaceRoot(newRoot interface{}) Aggregator[T, R] {
	a.addStage(StageReplaceRoot, "$replaceRoot", bson.D{{Key: "newRoot", Value: newRoot}}, false)
	return a
}

// ReplaceWith adds a $replaceWith stage
func (a *aggregator[T, R]) ReplaceWith(expression interface{}) Aggregator[T, R] {
	a.addStage(StageReplaceWith, "$replaceWith", expression, false)
	return a
}

// Sample adds a $sample stage
func (a *aggregator[T, R]) Sample(size int64) Aggregator[T, R] {
	a.addStage(StageSample, "$sample", bson.D{{Key: "size", Value: size}}, true)
	return a
}

// Unset adds a $unset stage
func (a *aggregator[T, R]) Unset(fields ...string) Aggregator[T, R] {
	a.addStage(StageUnset, "$unset", fields, false)
	return a
}

// Redact adds a $redact stage
func (a *aggregator[T, R]) Redact(expression interface{}) Aggregator[T, R] {
	a.addStage(StageRedact, "$redact", expression, false)
	return a
}

// GeoNear adds a $geoNear stage
func (a *aggregator[T, R]) GeoNear(coordinates [2]float64, options interface{}) Aggregator[T, R] {
	geoNearDoc := bson.M{
		"near": bson.M{
			"type":        "Point",
			"coordinates": coordinates,
		},
	}

	// Add options
	if optMap, ok := options.(map[string]interface{}); ok {
		for k, v := range optMap {
			geoNearDoc[k] = v
		}
	}

	a.addStage(StageGeoNear, "$geoNear", geoNearDoc, true)
	return a
}

// Exec executes the aggregation pipeline and returns all results
func (a *aggregator[T, R]) Exec() ([]*R, error) {
	defer a.Close()
	a.logger.Info("[ MONGO ] - Executing aggregation pipeline with %d stages", len(a.pipeline))

	opts := options.Aggregate().SetAllowDiskUse(a.allowDiskUse)
	cursor, err := a.collection.Aggregate(a.context, a.pipeline, opts)
	if err != nil {
		a.logger.Error("[ MONGO ] - Error executing aggregation: %v", err)
		return nil, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(a.context)

	var results []*R
	for cursor.Next(a.context) {
		var result R
		if err := cursor.Decode(&result); err != nil {
			a.logger.Error("[ MONGO ] - Error decoding aggregation result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		results = append(results, &result)
	}

	if err := cursor.Err(); err != nil {
		a.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	a.logger.Info("[ MONGO ] - Aggregation executed successfully, retrieved %d results", len(results))
	return results, nil
}

// ExecOne executes the aggregation pipeline and returns the first result
func (a *aggregator[T, R]) ExecOne() (*R, error) {
	defer a.Close()
	a.logger.Info("[ MONGO ] - Executing aggregation pipeline for single result")

	// Make a copy of the pipeline
	pipeline := make([]bson.D, len(a.pipeline))
	copy(pipeline, a.pipeline)

	// Check if there's already a limit stage
	if _, hasLimit := a.stageMap[StageLimit]; !hasLimit {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: 1}})
	}

	opts := options.Aggregate().SetAllowDiskUse(a.allowDiskUse)
	cursor, err := a.collection.Aggregate(a.context, pipeline, opts)
	if err != nil {
		a.logger.Error("[ MONGO ] - Error executing aggregation for single result: %v", err)
		return nil, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(a.context)

	if !cursor.Next(a.context) {
		if err := cursor.Err(); err != nil {
			a.logger.Error("[ MONGO ] - Cursor error: %v", err)
			return nil, fmt.Errorf("cursor error: %w", err)
		}
		a.logger.Info("[ MONGO ] - No results found")
		return nil, mongo.ErrNoDocuments
	}

	var result R
	if err := cursor.Decode(&result); err != nil {
		a.logger.Error("[ MONGO ] - Error decoding single result: %v", err)
		return nil, fmt.Errorf("error decoding result: %w", err)
	}

	a.logger.Info("[ MONGO ] - Single result aggregation executed successfully")
	return &result, nil
}

// ExecPaginated executes the aggregation pipeline with pagination
func (a *aggregator[T, R]) ExecPaginated(page, limit int64) ([]*R, error) {
	defer a.Close()
	a.logger.Info("[ MONGO ] - Executing paginated aggregation: page %d, limit %d", page, limit)

	// Make a copy of the pipeline
	pipeline := make([]bson.D, len(a.pipeline))
	copy(pipeline, a.pipeline)

	// Add pagination stages, replacing existing ones if they exist
	skip := (page - 1) * limit

	// Check if there are existing pagination stages
	skipPos, hasSkip := a.stageMap[StageSkip]
	limitPos, hasLimit := a.stageMap[StageLimit]

	if hasSkip {
		// Replace the skip stage
		pipeline[skipPos] = bson.D{{Key: "$skip", Value: skip}}
	} else {
		// Add a new skip stage
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: skip}})
	}

	if hasLimit {
		// Replace the limit stage
		pipeline[limitPos] = bson.D{{Key: "$limit", Value: limit}}
	} else {
		// Add a new limit stage
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	opts := options.Aggregate().SetAllowDiskUse(a.allowDiskUse)
	cursor, err := a.collection.Aggregate(a.context, pipeline, opts)
	if err != nil {
		a.logger.Error("[ MONGO ] - Error executing paginated aggregation: %v", err)
		return nil, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(a.context)

	var results []*R
	for cursor.Next(a.context) {
		var result R
		if err := cursor.Decode(&result); err != nil {
			a.logger.Error("[ MONGO ] - Error decoding paginated result: %v", err)
			return nil, fmt.Errorf("error decoding result: %w", err)
		}
		results = append(results, &result)
	}

	if err := cursor.Err(); err != nil {
		a.logger.Error("[ MONGO ] - Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	a.logger.Info("[ MONGO ] - Paginated aggregation executed successfully, retrieved %d results", len(results))
	return results, nil
}

// ExecCount executes the aggregation pipeline and returns the count of results
func (a *aggregator[T, R]) ExecCount() (int64, error) {
	defer a.Close()
	a.logger.Info("[ MONGO ] - Executing count aggregation")

	// Make a copy of the pipeline
	pipeline := make([]bson.D, len(a.pipeline))
	copy(pipeline, a.pipeline)

	// Add a count stage at the end if one doesn't already exist
	if _, hasCount := a.stageMap[StageCount]; !hasCount {
		pipeline = append(pipeline, bson.D{{Key: "$count", Value: "count"}})
	}

	opts := options.Aggregate().SetAllowDiskUse(a.allowDiskUse)
	cursor, err := a.collection.Aggregate(a.context, pipeline, opts)
	if err != nil {
		a.logger.Error("[ MONGO ] - Error executing count aggregation: %v", err)
		return 0, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(a.context)

	if !cursor.Next(a.context) {
		if err := cursor.Err(); err != nil {
			a.logger.Error("[ MONGO ] - Cursor error: %v", err)
			return 0, fmt.Errorf("cursor error: %w", err)
		}
		a.logger.Info("[ MONGO ] - No results found for count")
		return 0, nil
	}

	var result struct {
		Count int64 `bson:"count"`
	}
	if err := cursor.Decode(&result); err != nil {
		a.logger.Error("[ MONGO ] - Error decoding count result: %v", err)
		return 0, fmt.Errorf("error decoding result: %w", err)
	}

	a.logger.Info("[ MONGO ] - Count aggregation executed successfully, count: %d", result.Count)
	return result.Count, nil
}

// ExecRaw executes the aggregation pipeline and returns raw BSON results
func (a *aggregator[T, R]) ExecRaw() (interface{}, error) {
	defer a.Close()
	a.logger.Info("[ MONGO ] - Executing raw aggregation")

	opts := options.Aggregate().SetAllowDiskUse(a.allowDiskUse)
	cursor, err := a.collection.Aggregate(a.context, a.pipeline, opts)
	if err != nil {
		a.logger.Error("[ MONGO ] - Error executing raw aggregation: %v", err)
		return nil, fmt.Errorf("error executing aggregation: %w", err)
	}
	defer cursor.Close(a.context)

	var results []bson.M
	if err = cursor.All(a.context, &results); err != nil {
		a.logger.Error("[ MONGO ] - Error retrieving raw results: %v", err)
		return nil, fmt.Errorf("error retrieving results: %w", err)
	}

	a.logger.Info("[ MONGO ] - Raw aggregation executed successfully, retrieved %d results", len(results))
	return results, nil
}
