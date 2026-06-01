package functions

import (
	"context"
	"fmt"
	"net/http"

	"blog-cms-v2/src/libs"

	"github.com/aws/aws-lambda-go/events"
)

func RecommendComment(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slug := event.PathParameters["slug"]
	commentID := event.PathParameters["commentId"]

	if slug == "" || commentID == "" {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("slug and commentId are required")), nil
	}

	repo, err := libs.NewCommentRepository(ctx)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, err), nil
	}

	if err := repo.IncrementRecommend(ctx, slug, commentID); err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, fmt.Errorf("recommending: %w", err)), nil
	}

	return libs.Response(http.StatusOK, map[string]string{"status": "ok"}), nil
}
