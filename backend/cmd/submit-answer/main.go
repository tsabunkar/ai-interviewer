package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

type SubmitResponse struct {
	SubmissionID string `json:"submissionId"`
	Status       string `json:"status"`
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

	var submitReq models.SubmitRequest
	err := json.Unmarshal([]byte(req.Body), &submitReq)
	if err != nil {
		log.Printf("Failed to unmarshal request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       `{"error":"invalid request payload"}`,
		}, nil
	}

	// Validate inputs
	if submitReq.UserID == "" || submitReq.QuestionDate == "" || len(submitReq.Answer) < 100 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       `{"error":"validation failed: userId and questionDate are required; answer must be at least 100 characters"}`,
		}, nil
	}

	submissionID := uuid.New().String()

	dynamoClient, err := db.NewDynamoClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to initialize database client"}`,
		}, nil
	}

	// Save initial pending submission in DynamoDB
	submission := &models.Submission{
		SubmissionID: submissionID,
		UserID:       submitReq.UserID,
		QuestionDate: submitReq.QuestionDate,
		Answer:       submitReq.Answer,
		Status:       "pending",
		SubmittedAt:  time.Now().Format(time.RFC3339),
	}

	err = dynamoClient.PutSubmission(ctx, submission)
	if err != nil {
		log.Printf("Error writing pending submission: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to save initial submission"}`,
		}, nil
	}

	// Send message to SQS for processing
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading SQS config: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to configure queue client"}`,
		}, nil
	}

	sqsClient := sqs.NewFromConfig(cfg)
	queueURL := os.Getenv("SUBMISSIONS_QUEUE_URL")
	if queueURL == "" {
		log.Println("Warning: SUBMISSIONS_QUEUE_URL environment variable is empty!")
	}

	msgBody := models.SQSSubmissionMessage{
		SubmissionID: submissionID,
		UserID:       submitReq.UserID,
		QuestionDate: submitReq.QuestionDate,
		Answer:       submitReq.Answer,
	}

	msgBytes, err := json.Marshal(msgBody)
	if err != nil {
		log.Printf("Error marshalling SQS message: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to format queue message"}`,
		}, nil
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(msgBytes)),
	})
	if err != nil {
		log.Printf("Error sending message to SQS: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to enqueue submission"}`,
		}, nil
	}

	response := SubmitResponse{
		SubmissionID: submissionID,
		Status:       "pending",
	}

	responseBytes, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusAccepted,
		Headers:    headers,
		Body:       string(responseBytes),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
