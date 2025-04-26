package postgres

import "github.com/lib/pq"

func IsNoRowsFoundError(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "sql: no rows in result set" {
		return true
	}
	return false
}

// IsUniqueViolationError checks if the error is due to unique constraint violation
func IsUniqueViolationError(err error) bool {
	if err == nil {
		return false
	}
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505" // PostgreSQL unique violation code
}

// IsForeignKeyViolationError checks if the error is due to foreign key constraint violation
func IsForeignKeyViolationError(err error) bool {
	if err == nil {
		return false
	}
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23503" // PostgreSQL foreign key violation code
}