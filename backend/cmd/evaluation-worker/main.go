package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ai-interviewer/backend/internal/bedrock"
	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type SQSBatchItemFailure struct {
	ItemIdentifier string `json:"itemIdentifier"`
}

type SQSBatchResponse struct {
	BatchItemFailures []SQSBatchItemFailure `json:"batchItemFailures"`
}

func handleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) (SQSBatchResponse, error) {
	var failures []SQSBatchItemFailure

	dynamoClient, err := db.NewDynamoClient(ctx)
	if err != nil {
		log.Printf("Error initializing DynamoDB client: %v", err)
		// Mark all messages in batch as failed
		for _, record := range sqsEvent.Records {
			failures = append(failures, SQSBatchItemFailure{ItemIdentifier: record.MessageId})
		}
		return SQSBatchResponse{BatchItemFailures: failures}, nil
	}

	bedrockClient, err := bedrock.NewBedrockClient(ctx)
	if err != nil {
		log.Printf("Error initializing Bedrock client: %v", err)
		for _, record := range sqsEvent.Records {
			failures = append(failures, SQSBatchItemFailure{ItemIdentifier: record.MessageId})
		}
		return SQSBatchResponse{BatchItemFailures: failures}, nil
	}

	for _, record := range sqsEvent.Records {
		err := processMessage(ctx, record, dynamoClient, bedrockClient)
		if err != nil {
			log.Printf("Failed to process SQS message %s: %v", record.MessageId, err)
			failures = append(failures, SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
		}
	}

	return SQSBatchResponse{BatchItemFailures: failures}, nil
}

func processMessage(ctx context.Context, record events.SQSMessage, dynamoClient *db.DynamoDBClient, bedrockClient *bedrock.BedrockClient) error {
	var msg models.SQSSubmissionMessage
	err := json.Unmarshal([]byte(record.Body), &msg)
	if err != nil {
		log.Printf("Error unmarshalling SQS record body: %v", err)
		return err
	}

	log.Printf("Processing submission %s for user %s on date %s", msg.SubmissionID, msg.UserID, msg.QuestionDate)

	// 1. Fetch Today's Question Details
	question, err := dynamoClient.GetQuestion(ctx, msg.QuestionDate)
	if err != nil {
		log.Printf("Error fetching question for %s: %v", msg.QuestionDate, err)
		return err
	}

	questionTitle := "Unknown Question"
	questionDesc := ""
	if question != nil {
		questionTitle = question.Title
		questionDesc = question.Description
	} else {
		log.Printf("Warning: Question not found for date %s. Evaluating without description.", msg.QuestionDate)
	}

	// 2. Evaluate Answer with Bedrock
	log.Printf("Invoking Bedrock evaluation for submission %s...", msg.SubmissionID)
	evalResult, err := bedrockClient.EvaluateAnswer(ctx, questionDesc, msg.Answer)
	if err != nil {
		log.Printf("Error in Bedrock evaluation: %v", err)
		return err
	}

	log.Printf("Evaluation score for submission %s: %d", msg.SubmissionID, evalResult.Score)

	// 3. Update Submission in DynamoDB with Evaluation results
	err = dynamoClient.UpdateSubmissionResult(ctx, msg.SubmissionID, evalResult, evalResult.Score)
	if err != nil {
		log.Printf("Error updating submission result in DynamoDB: %v", err)
		return err
	}

	// 4. Update Leaderboard Entry
	log.Printf("Updating leaderboard for date %s, score %d", msg.QuestionDate, evalResult.Score)
	err = dynamoClient.UpdateLeaderboard(ctx, msg.QuestionDate, evalResult.Score, msg.UserID, questionTitle)
	if err != nil {
		log.Printf("Error updating leaderboard: %v", err)
		// We log the error but don't fail the message processing, since the evaluation itself was successful and saved.
	}

	log.Printf("Successfully processed submission %s", msg.SubmissionID)
	return nil
}

func main() {
	lambda.Start(handleSQSEvent)
}
