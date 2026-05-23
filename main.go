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
	default:
		lambda.Start(functions.GetPostBySlug)
	}
}
