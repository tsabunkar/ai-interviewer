package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LeaderboardResponse struct {
	Leaderboard    *models.LeaderboardEntry `json:"leaderboard"`
	TopSubmissions []models.Submission      `json:"topSubmissions"`
}

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

	date := req.QueryStringParameters["date"]
	if date == "" {
		loc := time.FixedZone("IST", 19800)
		date = time.Now().In(loc).Format("2006-01-02")
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

	// 1. Fetch Aggregated Leaderboard stats
	leaderboardEntry, err := dynamoClient.GetLeaderboard(ctx, date)
	if err != nil {
		log.Printf("Error fetching leaderboard entry: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to fetch leaderboard stats"}`,
		}, nil
	}

	// 2. Query Top 10 Submissions (GSI sorting score DESC)
	topSubmissions, err := dynamoClient.QueryTopSubmissions(ctx, date, 10)
	if err != nil {
		log.Printf("Error querying top submissions: %v", err)
		// We still return 200 with leaderboard stats even if querying detailed top list fails
		topSubmissions = []models.Submission{}
	}

	// If no entry exists yet, initialize a blank leaderboard entry
	if leaderboardEntry == nil {
		leaderboardEntry = &models.LeaderboardEntry{
			Date:             date,
			HighestScore:     0,
			TopScorer:        "None",
			TotalSubmissions: 0,
			QuestionTitle:    "No active question",
		}
	}

	response := LeaderboardResponse{
		Leaderboard:    leaderboardEntry,
		TopSubmissions: topSubmissions,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling leaderboard response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to format leaderboard response"}`,
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
