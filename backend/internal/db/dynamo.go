package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBClient struct {
	Client           *dynamodb.Client
	QuestionsTable   string
	SubmissionsTable string
	LeaderboardTable string
}

func NewDynamoClient(ctx context.Context) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	var svcOpts []func(*dynamodb.Options)
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		svcOpts = append(svcOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	client := dynamodb.NewFromConfig(cfg, svcOpts...)

	questionsTable := os.Getenv("QUESTIONS_TABLE")
	submissionsTable := os.Getenv("SUBMISSIONS_TABLE")
	leaderboardTable := os.Getenv("LEADERBOARD_TABLE")

	if questionsTable == "" {
		questionsTable = "ai-interviewer-questions"
	}
	if submissionsTable == "" {
		submissionsTable = "ai-interviewer-submissions"
	}
	if leaderboardTable == "" {
		leaderboardTable = "ai-interviewer-leaderboard"
	}

	return &DynamoDBClient{
		Client:           client,
		QuestionsTable:   questionsTable,
		SubmissionsTable: submissionsTable,
		LeaderboardTable: leaderboardTable,
	}, nil
}

// Questions Table Operations
func (d *DynamoDBClient) GetQuestion(ctx context.Context, date string) (*models.Question, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.QuestionsTable),
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
	}

	result, err := d.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Not found is not an error
	}

	var question models.Question
	err = attributevalue.UnmarshalMap(result.Item, &question)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal question: %w", err)
	}

	return &question, nil
}

func (d *DynamoDBClient) PutQuestion(ctx context.Context, question *models.Question) error {
	av, err := attributevalue.MarshalMap(question)
	if err != nil {
		return fmt.Errorf("failed to marshal question: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.QuestionsTable),
		Item:      av,
	}

	_, err = d.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put question: %w", err)
	}

	return nil
}

// Submissions Table Operations
func (d *DynamoDBClient) PutSubmission(ctx context.Context, submission *models.Submission) error {
	av, err := attributevalue.MarshalMap(submission)
	if err != nil {
		return fmt.Errorf("failed to marshal submission: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.SubmissionsTable),
		Item:      av,
	}

	_, err = d.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put submission: %w", err)
	}

	return nil
}

func (d *DynamoDBClient) GetSubmission(ctx context.Context, submissionId string) (*models.Submission, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.SubmissionsTable),
		Key: map[string]types.AttributeValue{
			"submissionId": &types.AttributeValueMemberS{Value: submissionId},
		},
	}

	result, err := d.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var submission models.Submission
	err = attributevalue.UnmarshalMap(result.Item, &submission)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal submission: %w", err)
	}

	return &submission, nil
}

func (d *DynamoDBClient) UpdateSubmissionResult(ctx context.Context, submissionId string, eval *models.EvaluationResult, score int) error {
	evalAv, err := attributevalue.Marshal(eval)
	if err != nil {
		return fmt.Errorf("failed to marshal evaluation: %w", err)
	}

	now := time.Now().Format(time.RFC3339)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.SubmissionsTable),
		Key: map[string]types.AttributeValue{
			"submissionId": &types.AttributeValueMemberS{Value: submissionId},
		},
		UpdateExpression: aws.String("SET #status = :status, score = :score, evaluation = :eval, evaluatedAt = :evaluatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":      &types.AttributeValueMemberS{Value: "evaluated"},
			":score":       &types.AttributeValueMemberN{Value: strconv.Itoa(score)},
			":eval":        evalAv,
			":evaluatedAt": &types.AttributeValueMemberS{Value: now},
		},
	}

	_, err = d.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update submission result: %w", err)
	}

	return nil
}

func (d *DynamoDBClient) QuerySubmissionsByUser(ctx context.Context, userId string) ([]models.Submission, error) {
	keyCond := expression.Key("userId").Equal(expression.Value(userId))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.SubmissionsTable),
		IndexName:                 aws.String("userId-date-index"),
		KeyConditionExpression:   expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	var submissions []models.Submission
	paginator := dynamodb.NewQueryPaginator(d.Client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query user submissions: %w", err)
		}

		var pageSubs []models.Submission
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageSubs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal page: %w", err)
		}
		submissions = append(submissions, pageSubs...)
	}

	return submissions, nil
}

func (d *DynamoDBClient) QueryTopSubmissions(ctx context.Context, date string, limit int32) ([]models.Submission, error) {
	keyCond := expression.Key("questionDate").Equal(expression.Value(date))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.SubmissionsTable),
		IndexName:                 aws.String("questionDate-score-index"),
		KeyConditionExpression:   expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // Descending order (highest score first)
		Limit:                     aws.Int32(limit),
	}

	result, err := d.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query top submissions: %w", err)
	}

	var submissions []models.Submission
	err = attributevalue.UnmarshalListOfMaps(result.Items, &submissions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal top submissions: %w", err)
	}

	return submissions, nil
}

// Leaderboard Table Operations
func (d *DynamoDBClient) GetLeaderboard(ctx context.Context, date string) (*models.LeaderboardEntry, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.LeaderboardTable),
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
	}

	result, err := d.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard entry: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var entry models.LeaderboardEntry
	err = attributevalue.UnmarshalMap(result.Item, &entry)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaderboard: %w", err)
	}

	return &entry, nil
}

func (d *DynamoDBClient) UpdateLeaderboard(ctx context.Context, date string, score int, userId string, questionTitle string) error {
	entry, err := d.GetLeaderboard(ctx, date)
	if err != nil {
		return fmt.Errorf("failed to get current leaderboard for update: %w", err)
	}

	if entry == nil {
		newEntry := &models.LeaderboardEntry{
			Date:             date,
			HighestScore:     score,
			TopScorer:        userId,
			TotalSubmissions: 1,
			QuestionTitle:    questionTitle,
		}
		av, err := attributevalue.MarshalMap(newEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal new leaderboard entry: %w", err)
		}
		_, err = d.Client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(d.LeaderboardTable),
			Item:      av,
		})
		if err != nil {
			return fmt.Errorf("failed to insert new leaderboard entry: %w", err)
		}
		return nil
	}

	updateExpr := "SET totalSubmissions = totalSubmissions + :one, questionTitle = :qtitle"
	exprValues := map[string]types.AttributeValue{
		":one":    &types.AttributeValueMemberN{Value: "1"},
		":qtitle": &types.AttributeValueMemberS{Value: questionTitle},
	}

	if score > entry.HighestScore {
		updateExpr += ", highestScore = :score, topScorer = :userId"
		exprValues[":score"] = &types.AttributeValueMemberN{Value: strconv.Itoa(score)}
		exprValues[":userId"] = &types.AttributeValueMemberS{Value: userId}
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.LeaderboardTable),
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprValues,
	}

	_, err = d.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	return nil
}
