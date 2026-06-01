package libs

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"blog-cms-v2/src/definitions"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CommentRepository abstracts storage (DynamoDB in Lambda, in-memory locally).
type CommentRepository interface {
	GetComments(ctx context.Context, slug string) ([]definitions.Comment, error)
	PutComment(ctx context.Context, c definitions.Comment) error
	IncrementRecommend(ctx context.Context, slug, commentID string) error
}

func NewCommentRepository(ctx context.Context) (CommentRepository, error) {
	if os.Getenv("COMMENTS_TABLE") != "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading AWS config: %w", err)
		}
		table := os.Getenv("COMMENTS_TABLE")
		return &dynamoRepo{client: dynamodb.NewFromConfig(cfg), table: table}, nil
	}
	return &memRepo{}, nil
}

// ── DynamoDB ──────────────────────────────────────────────────────────────────

type dynamoRepo struct {
	client *dynamodb.Client
	table  string
}

func (r *dynamoRepo) GetComments(ctx context.Context, slug string) ([]definitions.Comment, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String("#slug = :slug"),
		FilterExpression:       aws.String("#active = :true"),
		ExpressionAttributeNames: map[string]string{
			"#slug":   "slug",
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":slug": &types.AttributeValueMemberS{Value: slug},
			":true": &types.AttributeValueMemberBOOL{Value: true},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("querying comments: %w", err)
	}
	var comments []definitions.Comment
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *dynamoRepo) PutComment(ctx context.Context, c definitions.Comment) error {
	item, err := attributevalue.MarshalMap(c)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	return err
}

func (r *dynamoRepo) IncrementRecommend(ctx context.Context, slug, commentID string) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"slug":      &types.AttributeValueMemberS{Value: slug},
			"commentId": &types.AttributeValueMemberS{Value: commentID},
		},
		UpdateExpression: aws.String("ADD recommends :one"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":one": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	return err
}

// ── In-memory (local dev) ─────────────────────────────────────────────────────

type memRepo struct {
	mu       sync.RWMutex
	comments []definitions.Comment
}

func (r *memRepo) GetComments(_ context.Context, slug string) ([]definitions.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []definitions.Comment
	for _, c := range r.comments {
		if c.Slug == slug && c.Active {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *memRepo) PutComment(_ context.Context, c definitions.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comments = append(r.comments, c)
	return nil
}

func (r *memRepo) IncrementRecommend(_ context.Context, slug, commentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.comments {
		if r.comments[i].Slug == slug && r.comments[i].CommentID == commentID {
			r.comments[i].Recommends++
			return nil
		}
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func NewCommentID() string {
	ts := strconv.FormatInt(time.Now().UnixNano(), 36)
	rnd := strconv.FormatInt(rand.Int63n(1<<32), 36)
	return ts + "-" + rnd
}
