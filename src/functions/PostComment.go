package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blog-cms-v2/src/definitions"
	"blog-cms-v2/src/libs"

	"github.com/aws/aws-lambda-go/events"
)

type commentRequest struct {
	Author   string `json:"author"`
	Text     string `json:"text"`
	ParentID string `json:"parentId"`
}

func PostComment(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slug := event.PathParameters["slug"]
	if slug == "" {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("slug is required")), nil
	}

	var req commentRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("invalid body: %w", err)), nil
	}

	req.Author = strings.TrimSpace(req.Author)
	req.Text = strings.TrimSpace(req.Text)

	if req.Author == "" || req.Text == "" {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("author and text are required")), nil
	}

	repo, err := libs.NewCommentRepository(ctx)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, err), nil
	}

	c := definitions.Comment{
		Slug:       slug,
		CommentID:  libs.NewCommentID(),
		Author:     req.Author,
		Text:       req.Text,
		CreatedAt:  time.Now().UTC(),
		Active:     false,
		ParentID:   req.ParentID,
		Recommends: 0,
	}

	if err := repo.PutComment(ctx, c); err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, fmt.Errorf("saving comment: %w", err)), nil
	}

	return libs.Response(http.StatusCreated, map[string]string{"status": "pending"}), nil
}
