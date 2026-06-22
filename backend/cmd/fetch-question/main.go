package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// CORS Headers
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Amz-Date,X-Api-Key",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Content-Type":                 "application/json",
	}

	// If OPTIONS preflight
	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	// Get date in IST
	loc := time.FixedZone("IST", 19800)
	today := time.Now().In(loc).Format("2006-01-02")
	log.Printf("Fetching question for date: %s", today)

	dynamoClient, err := db.NewDynamoClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to initialize database client"}`,
		}, nil
	}

	question, err := dynamoClient.GetQuestion(ctx, today)
	if err != nil {
		log.Printf("Error fetching question: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to fetch question from database"}`,
		}, nil
	}

	if question == nil {
		log.Printf("No question generated yet for today (%s)", today)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Headers:    headers,
			Body:       `{"error":"no challenge active for today yet"}`,
		}, nil
	}

	responseBody, err := json.Marshal(question)
	if err != nil {
		log.Printf("Error marshalling question response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to format question response"}`,
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
