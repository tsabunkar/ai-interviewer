package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ai-interviewer/backend/internal/bedrock"
	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
	"github.com/ai-interviewer/backend/internal/questions"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
)

func handleScheduledEvent(ctx context.Context, event json.RawMessage) error {
	log.Printf("Received EventBridge schedule event: %s", string(event))

	// IST TimeZone Calculation
	loc := time.FixedZone("IST", 19800) // 5.5 * 3600 = 19800
	nowIST := time.Now().In(loc)
	today := nowIST.Format("2006-01-02")
	log.Printf("Generating question for date: %s", today)

	dynamoClient, err := db.NewDynamoClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		return err
	}

	// Check if today's question already exists
	existing, err := dynamoClient.GetQuestion(ctx, today)
	if err != nil {
		log.Printf("Error checking existing question for %s: %v", today, err)
		return err
	}

	if existing != nil {
		log.Printf("Question already exists for %s: %s", today, existing.Title)
		return nil
	}

	var question *models.Question

	// Initialize Bedrock client and attempt generation
	bedrockClient, err := bedrock.NewBedrockClient(ctx)
	if err == nil {
		log.Println("Attempting to generate question using Bedrock Claude...")
		question, err = bedrockClient.GenerateQuestion(ctx)
		if err != nil {
			log.Printf("Bedrock question generation failed: %v", err)
		}
	} else {
		log.Printf("Failed to initialize Bedrock client: %v", err)
	}

	// Fallback to curated pool if Bedrock failed or was not initialized
	if question == nil {
		log.Println("Falling back to curated question pool...")
		fallbackQ := questions.GetRandomQuestion(today)
		question = &fallbackQ
	}

	// Finalize question fields
	question.Date = today
	question.QuestionID = uuid.New().String()
	question.CreatedAt = nowIST.Format(time.RFC3339)

	// Save to Questions Table
	err = dynamoClient.PutQuestion(ctx, question)
	if err != nil {
		log.Printf("Failed to save today's question to DynamoDB: %v", err)
		return err
	}

	log.Printf("Successfully generated and saved question for %s: %s", today, question.Title)
	return nil
}

func main() {
	lambda.Start(handleScheduledEvent)
}
