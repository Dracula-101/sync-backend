package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync-backend/utils"
	"time"
)

type Query[T any] interface {
	Close()
	FindOne(query string, args ...any) (*T, error)
	FindAll(query string, args ...any) ([]*T, error)
	FindPaginated(query string, page int64, limit int64, args ...any) ([]*T, error)
	InsertOne(query string, args ...any) (int64, error)
	InsertAndRetrieveOne(insertQuery string, retrieveQuery string, args ...any) (*T, error)
	InsertMany(query string, argsList [][]any) ([]int64, error)
	InsertAndRetrieveMany(insertQuery string, retrieveQuery string, argsList [][]any) ([]*T, error)
	FilterOne(query string, args ...any) (*T, error)
	FilterMany(query string, args ...any) ([]*T, error)
	FilterPaginated(query string, page int64, limit int64, args ...any) ([]*T, error)
	FilterCount(query string, args ...any) (int64, error)
	UpdateOne(query string, args ...any) (int64, error)
	UpdateMany(query string, args ...any) (int64, error)
	DeleteOne(query string, args ...any) (int64, error)
	DeleteMany(query string, args ...any) (int64, error)
	ExecContext(query string, args ...any) (sql.Result, error)
	QueryContext(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(query string, args ...any) *sql.Row
}

type query[T any] struct {
	logger  utils.AppLogger
	db      *sql.DB
	context context.Context
	cancel  context.CancelFunc
}

func newSingleQuery[T any](logger utils.AppLogger, db *sql.DB, timeout time.Duration) Query[T] {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &query[T]{
		logger:  logger,
		context: ctx,
		cancel:  cancel,
		db:      db,
	}
}

func newQuery[T any](logger utils.AppLogger, context context.Context, db *sql.DB) Query[T] {
	return &query[T]{
		logger:  logger,
		context: context,
		db:      db,
	}
}

func (q *query[T]) Close() {
	if q.cancel != nil {
		q.cancel()
	}
}

func (q *query[T]) scanRow(row *sql.Row) (*T, error) {
	var result T

	resultValue := reflect.ValueOf(&result).Elem()

	numFields := resultValue.NumField()

	// Create a slice to hold pointers to each field
	fieldPtrs := make([]interface{}, numFields)
	for i := 0; i < numFields; i++ {
		fieldPtrs[i] = resultValue.Field(i).Addr().Interface()
	}

	// Scan directly into each field pointer
	if err := row.Scan(fieldPtrs...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &result, nil
}

func (q *query[T]) scanRows(rows *sql.Rows) ([]*T, error) {
	var results []*T
	defer rows.Close()

	// Get column names from the query result
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting column names: %w", err)
	}

	// Create a template value to determine field count and types
	var templateValue T
	resultType := reflect.TypeOf(templateValue)

	for rows.Next() {
		var result T
		resultValue := reflect.ValueOf(&result).Elem()

		// Create a slice of interface{} to hold values
		fieldPtrs := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			// Find the struct field that corresponds to this column
			// (This is simplified; in a real implementation you might need field tags to map columns to fields)
			if i < resultType.NumField() {
				fieldPtrs[i] = resultValue.Field(i).Addr().Interface()
			} else {
				// Handle case when database returns more columns than struct fields
				var placeholder interface{}
				fieldPtrs[i] = &placeholder
			}
		}

		// Scan into the pointers
		if err := rows.Scan(fieldPtrs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return results, nil
}

func (q *query[T]) ExecContext(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := q.db.ExecContext(q.context, query, args...)
	elapsed := time.Since(start)

	q.logger.Debug("ExecContext", map[string]interface{}{
		"query":    query,
		"args":     args,
		"duration": elapsed,
		"error":    err,
	})

	return result, err
}

func (q *query[T]) QueryContext(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := q.db.QueryContext(q.context, query, args...)
	elapsed := time.Since(start)

	q.logger.Debug("QueryContext", map[string]interface{}{
		"query":    query,
		"args":     args,
		"duration": elapsed,
		"error":    err,
	})

	return rows, err
}

func (q *query[T]) QueryRowContext(query string, args ...any) *sql.Row {
	start := time.Now()
	row := q.db.QueryRowContext(q.context, query, args...)
	elapsed := time.Since(start)

	q.logger.Debug("QueryRowContext", map[string]interface{}{
		"query":    query,
		"args":     args,
		"duration": elapsed,
	})

	return row
}

func (q *query[T]) FindOne(query string, args ...any) (*T, error) {
	row := q.QueryRowContext(query, args...)
	return q.scanRow(row)
}

func (q *query[T]) FindAll(query string, args ...any) ([]*T, error) {
	rows, err := q.QueryContext(query, args...)
	if err != nil {
		return nil, err
	}
	return q.scanRows(rows)
}

func (q *query[T]) FindPaginated(query string, page int64, limit int64, args ...any) ([]*T, error) {
	offset := (page - 1) * limit
	paginatedQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)
	return q.FindAll(paginatedQuery, args...)
}

func (q *query[T]) InsertOne(query string, args ...any) (int64, error) {
	result, err := q.ExecContext(query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (q *query[T]) InsertAndRetrieveOne(insertQuery string, retrieveQuery string, args ...any) (*T, error) {
	id, err := q.InsertOne(insertQuery, args...)
	if err != nil {
		return nil, err
	}
	return q.FindOne(retrieveQuery, id)
}

func (q *query[T]) InsertMany(query string, argsList [][]any) ([]int64, error) {
	ids := make([]int64, 0, len(argsList))

	for _, args := range argsList {
		id, err := q.InsertOne(query, args...)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (q *query[T]) InsertAndRetrieveMany(insertQuery string, retrieveQuery string, argsList [][]any) ([]*T, error) {
	ids, err := q.InsertMany(insertQuery, argsList)
	if err != nil {
		return nil, err
	}

	placeholders := ""
	idArgs := make([]any, 0, len(ids))

	for i, id := range ids {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		idArgs = append(idArgs, id)
	}

	fullRetrieveQuery := fmt.Sprintf(retrieveQuery, placeholders)
	return q.FindAll(fullRetrieveQuery, idArgs...)
}

func (q *query[T]) FilterOne(query string, args ...any) (*T, error) {
	return q.FindOne(query, args...)
}

func (q *query[T]) FilterMany(query string, args ...any) ([]*T, error) {
	return q.FindAll(query, args...)
}

func (q *query[T]) FilterPaginated(query string, page int64, limit int64, args ...any) ([]*T, error) {
	return q.FindPaginated(query, page, limit, args...)
}

func (q *query[T]) FilterCount(query string, args ...any) (int64, error) {
	var count int64
	err := q.QueryRowContext(query, args...).Scan(&count)
	return count, err
}

func (q *query[T]) UpdateOne(query string, args ...any) (int64, error) {
	result, err := q.ExecContext(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *query[T]) UpdateMany(query string, args ...any) (int64, error) {
	return q.UpdateOne(query, args...)
}

func (q *query[T]) DeleteOne(query string, args ...any) (int64, error) {
	result, err := q.ExecContext(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *query[T]) DeleteMany(query string, args ...any) (int64, error) {
	return q.DeleteOne(query, args...)
}

func NewPostgresQuery[T any](logger utils.AppLogger, db *sql.DB, timeout time.Duration) Query[T] {
	return newSingleQuery[T](logger, db, timeout)
}

func NewPostgresQueryWithContext[T any](logger utils.AppLogger, ctx context.Context, db *sql.DB) Query[T] {
	return newQuery[T](logger, ctx, db)
}
