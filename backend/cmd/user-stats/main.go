package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
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

	userID := req.PathParameters["userId"]
	if userID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       `{"error":"userId path parameter is required"}`,
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

	submissions, err := dynamoClient.QuerySubmissionsByUser(ctx, userID)
	if err != nil {
		log.Printf("Error querying user submissions: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to query user submissions"}`,
		}, nil
	}

	// Sort submissions descending by submittedAt time
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].SubmittedAt > submissions[j].SubmittedAt
	})

	totalSubmissions := len(submissions)
	var evaluatedCount int
	var scoreSum int

	// Streak calculation prep
	dateSet := make(map[string]bool)
	for _, sub := range submissions {
		dateSet[sub.QuestionDate] = true
		if sub.Status == "evaluated" {
			evaluatedCount++
			scoreSum += sub.Score
		}
	}

	var dates []string
	for d := range dateSet {
		dates = append(dates, d)
	}
	// Sort dates descending (most recent first)
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	currentStreak := 0
	longestStreak := 0

	// IST Date references
	loc := time.FixedZone("IST", 19800)
	todayStr := time.Now().In(loc).Format("2006-01-02")
	yesterdayStr := time.Now().In(loc).AddDate(0, 0, -1).Format("2006-01-02")

	if len(dates) > 0 {
		hasActivity := false
		startIndex := 0
		if dates[0] == todayStr {
			hasActivity = true
			startIndex = 0
		} else if dates[0] == yesterdayStr {
			hasActivity = true
			startIndex = 0
		}

		if hasActivity {
			currentStreak = 1
			for i := startIndex; i < len(dates)-1; i++ {
				currTime, _ := time.Parse("2006-01-02", dates[i])
				nextTime, _ := time.Parse("2006-01-02", dates[i+1])
				diffDays := int(currTime.Sub(nextTime).Hours() / 24)
				if diffDays == 1 {
					currentStreak++
				} else {
					break
				}
			}
		}

		// Calculate longest streak across all history
		tempStreak := 1
		for i := 0; i < len(dates)-1; i++ {
			currTime, _ := time.Parse("2006-01-02", dates[i])
			nextTime, _ := time.Parse("2006-01-02", dates[i+1])
			diffDays := int(currTime.Sub(nextTime).Hours() / 24)
			if diffDays == 1 {
				tempStreak++
			} else {
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
				tempStreak = 1
			}
		}
		if tempStreak > longestStreak {
			longestStreak = tempStreak
		}
	}

	averageScore := 0.0
	if evaluatedCount > 0 {
		averageScore = float64(scoreSum) / float64(evaluatedCount)
	}

	// Calculate weak areas: group scores by category
	categories := []string{"Requirements", "API Design", "Database", "Cache", "Scalability", "Availability", "Tradeoffs"}
	categorySums := make(map[string]int)
	categoryCounts := make(map[string]int)

	// Initialize all categories with 0.0 average
	for _, cat := range categories {
		categorySums[cat] = 0
		categoryCounts[cat] = 0
	}

	for _, sub := range submissions {
		if sub.Status == "evaluated" && sub.Evaluation != nil && sub.Evaluation.Categories != nil {
			for catName, catScoreObj := range sub.Evaluation.Categories {
				categorySums[catName] += catScoreObj.Score
				categoryCounts[catName]++
			}
		}
	}

	var weakAreas []models.WeakArea
	for _, cat := range categories {
		avg := 0.0
		if categoryCounts[cat] > 0 {
			avg = float64(categorySums[cat]) / float64(categoryCounts[cat])
		}
		weakAreas = append(weakAreas, models.WeakArea{
			Category:     cat,
			AverageScore: avg,
		})
	}

	// Sort weakAreas ascending by average score (lowest score first -> weakest area)
	sort.Slice(weakAreas, func(i, j int) bool {
		return weakAreas[i].AverageScore < weakAreas[j].AverageScore
	})

	// Build recentSubmissions list (max 20)
	var recentSummaries []models.SubmissionSummary
	limit := 20
	if len(submissions) < limit {
		limit = len(submissions)
	}
	for i := 0; i < limit; i++ {
		sub := submissions[i]
		recentSummaries = append(recentSummaries, models.SubmissionSummary{
			SubmissionID: sub.SubmissionID,
			QuestionDate: sub.QuestionDate,
			Score:        sub.Score,
			Status:       sub.Status,
			SubmittedAt:  sub.SubmittedAt,
		})
	}

	userStats := models.UserStats{
		UserID:            userID,
		TotalSubmissions:  totalSubmissions,
		AverageScore:      averageScore,
		CurrentStreak:     currentStreak,
		LongestStreak:     longestStreak,
		WeakAreas:         weakAreas,
		RecentSubmissions: recentSummaries,
	}

	responseBody, err := json.Marshal(userStats)
	if err != nil {
		log.Printf("Error marshalling user stats response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error":"failed to format stats response"}`,
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
