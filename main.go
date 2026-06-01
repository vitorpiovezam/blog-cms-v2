package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"blog-cms-v2/src/functions"
)

func main() {
	if os.Getenv("SERVE_LOCAL") == "true" {
		runHTTPServer()
		return
	}

	switch os.Getenv("AWS_LAMBDA_FUNCTION_NAME") {
	case "blog-cms-v2-dev-getAllPosts":
		lambda.Start(functions.GetAllPosts)
	case "blog-cms-v2-dev-getComments":
		lambda.Start(functions.GetComments)
	case "blog-cms-v2-dev-postComment":
		lambda.Start(functions.PostComment)
	case "blog-cms-v2-dev-recommendComment":
		lambda.Start(functions.RecommendComment)
	default:
		lambda.Start(functions.GetPostBySlug)
	}
}
