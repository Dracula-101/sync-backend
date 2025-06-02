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

func IsDuplicateKeyError(err error, key ...string) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "duplicate key error") {
		if len(key) == 0 {
			return true
		}
		if strings.Contains(err.Error(), key[0]) {
			return true
		}
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
