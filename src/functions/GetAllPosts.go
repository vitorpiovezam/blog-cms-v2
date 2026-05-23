package functions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"blog-cms-v2/src/libs"
)

func GetAllPosts(ctx context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	postService := libs.NewPostService()
	posts, err := postService.GetAllPosts(ctx)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, fmt.Errorf("fetching posts: %w", err)), nil
	}
	return libs.Response(http.StatusOK, posts), nil
}
