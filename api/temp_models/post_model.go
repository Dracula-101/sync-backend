package tempmodels

// import (
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Post struct {
// 	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	PostId    string             `bson:"postId" json:"postId"`
// 	AuthorId  string             `bson:"authorId" json:"authorId"`
// 	Title     string             `bson:"title" json:"title"`
// 	Content   string             `bson:"content" json:"content"`
// 	Images    []Image            `bson:"images" json:"images"`
// 	Video     Video              `bson:"video" json:"video"`
// 	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
// 	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
// 	DeletedAt primitive.DateTime `bson:"deletedAt" json:"deletedAt"`
// 	Comments  []Comment          `bson:"comments" json:"comments"`
// 	Tags      []string           `bson:"tags" json:"tags"`
// }

// type Video struct {
// 	Url    string `bson:"url" json:"url"`
// 	Width  int    `bson:"width" json:"width"`
// 	Height int    `bson:"height" json:"height"`
// }

// type Comment struct {
// 	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	CommentId  string             `bson:"commentId" json:"commentId"`
// 	AuthorId   string             `bson:"authorId" json:"authorId"`
// 	PostId     string             `bson:"postId" json:"postId"`
// 	ParentId   string             `bson:"parentId" json:"parentId"`
// 	Path       string             `bson:"path" json:"path"`
// 	Level      int                `bson:"level" json:"level"`
// 	ReplyTo    string             `bson:"replyTo" json:"replyTo"`
// 	Body       string             `bson:"body" json:"body"`
// 	Images     []Image            `bson:"images" json:"images"`
// 	Replies    []Comment          `bson:"replies" json:"replies"`
// 	Synergy    int                `bson:"synergy" json:"synergy"`
// 	ReplyCount int                `bson:"replyCount" json:"replyCount"`
// 	Edited     bool               `bson:"edited" json:"edited"`
// 	CreatedAt  primitive.DateTime `bson:"createdAt" json:"createdAt"`
// 	UpdatedAt  primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
// 	DeletedAt  primitive.DateTime `bson:"deletedAt" json:"deletedAt"`
// }
