package application

import (
	user "sync-backend/api/user/model"
	"sync-backend/arch/mongo"
)

func EnsureDbIndexes(db mongo.Database) {
	go mongo.Document[user.User](&user.User{}).EnsureIndexes(db)
}
