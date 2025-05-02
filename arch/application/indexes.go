package application

import (
	session "sync-backend/api/common/session/model"
	user "sync-backend/api/user/model"
	"sync-backend/arch/mongo"
)

func EnsureDbIndexes(db mongo.Database) {
	go mongo.Document[user.User](&user.User{}).EnsureIndexes(db)
	go mongo.Document[session.Session](&session.Session{}).EnsureIndexes(db)
}
