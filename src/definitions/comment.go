package definitions

import "time"

type Comment struct {
	Slug       string    `json:"slug" dynamodbav:"slug"`
	CommentID  string    `json:"commentId" dynamodbav:"commentId"`
	Author     string    `json:"author" dynamodbav:"author"`
	Text       string    `json:"text" dynamodbav:"text"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt"`
	Active     bool      `json:"active" dynamodbav:"active"`
	ParentID   string    `json:"parentId,omitempty" dynamodbav:"parentId,omitempty"`
	Recommends int       `json:"recommends" dynamodbav:"recommends"`
}
