package functions

import (
	"context"
	"fmt"
	"net/http"

	"blog-cms-v2/src/definitions"
	"blog-cms-v2/src/libs"

	"github.com/aws/aws-lambda-go/events"
)

func GetComments(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slug := event.PathParameters["slug"]
	if slug == "" {
		return libs.ErrResponse(http.StatusBadRequest, fmt.Errorf("slug is required")), nil
	}

	repo, err := libs.NewCommentRepository(ctx)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, err), nil
	}

	comments, err := repo.GetComments(ctx, slug)
	if err != nil {
		return libs.ErrResponse(http.StatusInternalServerError, fmt.Errorf("fetching comments: %w", err)), nil
	}

	if comments == nil {
		comments = []definitions.Comment{}
	}

	return libs.Response(http.StatusOK, comments), nil
}
