package libs

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
)

var corsHeaders = map[string]string{
	"Access-Control-Allow-Origin": "*",
	"Access-Control-Allow-Headers": "Content-Type",
	"Access-Control-Allow-Methods": "GET, OPTIONS",
}

func Response(statusCode int, body any) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body: string(b),
		Headers: corsHeaders,
	}
}

func ErrResponse(statusCode int, err error) events.APIGatewayProxyResponse {
	return Response(statusCode, map[string]string{"error": err.Error()})
}
