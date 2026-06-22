package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Amz-Date,X-Api-Key",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Content-Type":                 "application/json",
	}

	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	submissionID := req.PathParameters["submissionId"]
	if submissionID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       `{"error":"submissionId path parameter is required"}`,
		}, nil
	}

	dynamoClient, err := db.NewDynamoClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to initialize database client"}`,
		}, nil
	}

	submission, err := dynamoClient.GetSubmission(ctx, submissionID)
	if err != nil {
		log.Printf("Error fetching submission %s: %v", submissionID, err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to fetch submission details"}`,
		}, nil
	}

	if submission == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Headers:    headers,
			Body:       `{"error":"submission not found"}`,
		}, nil
	}

	responseBody, err := json.Marshal(submission)
	if err != nil {
		log.Printf("Error marshalling submission response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to format submission response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
