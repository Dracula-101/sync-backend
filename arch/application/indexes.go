package application

import (
	comment "sync-backend/api/comment/model"
	session "sync-backend/api/common/session/model"
	community "sync-backend/api/community/model"
	post "sync-backend/api/post/model"
	user "sync-backend/api/user/model"
	"sync-backend/arch/mongo"
)

// EnsureDbIndexes ensures all database indexes are created for all collections
func EnsureDbIndexes(db mongo.Database) {
	go mongo.Document[user.User](&user.User{}).EnsureIndexes(db)
	go mongo.Document[session.Session](&session.Session{}).EnsureIndexes(db)
	go mongo.Document[community.Community](&community.Community{}).EnsureIndexes(db)
	go mongo.Document[post.Post](&post.Post{}).EnsureIndexes(db)
	go mongo.Document[post.PostInteraction](&post.PostInteraction{}).EnsureIndexes(db)
	go mongo.Document[comment.Comment](&comment.Comment{}).EnsureIndexes(db)
}
