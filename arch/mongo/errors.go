package mongo

import (
	"strings"
)

func IsNoDocumentFoundError(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "duplicate key error") {
		return true
	}
	if strings.Contains(err.Error(), "E11000 duplicate key error") {
		return true
	}
	if strings.Contains(err.Error(), "MongoServerError") {
		return true
	}
	if strings.Contains(err.Error(), "DuplicateKey") {
		return true
	}
	if strings.Contains(err.Error(), "DuplicateKeyError") {
		return true
	}
	return false
}

func IsTransactionError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "TransactionAborted") {
		return true
	}
	if strings.Contains(err.Error(), "TransactionTooOld") {
		return true
	}
	if strings.Contains(err.Error(), "TransactionNotFound") {
		return true
	}
	return false
}
