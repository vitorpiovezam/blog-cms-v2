package functions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"blog-cms-v2/src/libs"
)

func GetPostBySlug(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slug := event.PathParameters["slug"]
	if slug == "" {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("slug is not defined")), nil
	}

	postService := libs.NewPostService()
	post, err := postService.GetPostBySlug(ctx, slug)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, fmt.Errorf("fetching post: %w", err)), nil
	}
	if post == nil {
		return libs.ErrResponse(http.StatusNotFound, fmt.Errorf("post not found: %s", slug)), nil
	}
	return libs.Response(http.StatusOK, post), nil
}
