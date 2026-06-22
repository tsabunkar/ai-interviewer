package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/ai-interviewer/backend/internal/db"
	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var (
	submissionEvals sync.Map // submissionId -> *models.Submission
	evalWorkers     sync.WaitGroup
)

func main() {
	ctx := context.Background()

	var dbClient *db.DynamoDBClient
	var lastErr error
	for i := 0; i < 30; i++ {
		dbClient, lastErr = db.NewDynamoClient(ctx)
		if lastErr == nil {
			break
		}
		log.Printf("Waiting for DynamoDB... (%d/30): %v", i+1, lastErr)
		time.Sleep(2 * time.Second)
	}
	if dbClient == nil {
		log.Fatalf("Failed to create DynamoDB client: %v", lastErr)
	}

	if err := initTables(ctx, dbClient); err != nil {
		log.Fatalf("Failed to initialize tables: %v", err)
	}

	if err := seedTodaysQuestion(ctx, dbClient); err != nil {
		log.Printf("Warning: failed to seed today's question: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(swaggerUIPage))
	})

	mux.HandleFunc("GET /swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	mux.HandleFunc("GET /question", handleGetQuestion(dbClient))
	mux.HandleFunc("POST /answer", handleSubmitAnswer(dbClient))
	mux.HandleFunc("GET /results/{submissionId}", handleGetResults(dbClient))
	mux.HandleFunc("GET /leaderboard", handleGetLeaderboard(dbClient))
	mux.HandleFunc("GET /stats/{userId}", handleGetUserStats(dbClient))

	handler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func initTables(ctx context.Context, client *db.DynamoDBClient) error {
	tables := []tableDef{
		{
			Name: client.QuestionsTable,
			PK:   "date",
		},
		{
			Name: client.SubmissionsTable,
			PK:   "submissionId",
			GSIs: []gsiDef{
				{Name: "userId-date-index", PK: "userId", SK: "questionDate", PKType: "S", SKType: "S"},
				{Name: "questionDate-score-index", PK: "questionDate", SK: "score", PKType: "S", SKType: "N"},
			},
		},
		{
			Name: client.LeaderboardTable,
			PK:   "date",
		},
	}

	for _, t := range tables {
		exists, err := tableExists(ctx, client.Client, t.Name)
		if err != nil {
			return fmt.Errorf("check table %s: %w", t.Name, err)
		}
		if exists {
			log.Printf("Table already exists: %s", t.Name)
			continue
		}

		if err := createTable(ctx, client.Client, t); err != nil {
			return fmt.Errorf("create table %s: %w", t.Name, err)
		}
		log.Printf("Created table: %s", t.Name)
	}

	return nil
}

func tableExists(ctx context.Context, svc *dynamodb.Client, name string) (bool, error) {
	_, err := svc.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type gsiDef struct {
	Name   string
	PK     string
	SK     string
	PKType string
	SKType string
}

type tableDef struct {
	Name string
	PK   string
	GSIs []gsiDef
}

func createTable(ctx context.Context, svc *dynamodb.Client, tbl tableDef) error {
	attrDefMap := map[string]types.ScalarAttributeType{}
	for _, gsi := range tbl.GSIs {
		attrDefMap[gsi.PK] = toAttrType(gsi.PKType)
		if gsi.SK != "" {
			attrDefMap[gsi.SK] = toAttrType(gsi.SKType)
		}
	}
	if _, ok := attrDefMap[tbl.PK]; !ok {
		attrDefMap[tbl.PK] = types.ScalarAttributeTypeS
	}

	attrDefs := make([]types.AttributeDefinition, 0, len(attrDefMap))
	for name, typ := range attrDefMap {
		attrDefs = append(attrDefs, types.AttributeDefinition{
			AttributeName: aws.String(name),
			AttributeType: typ,
		})
	}

	keySchema := []types.KeySchemaElement{
		{AttributeName: aws.String(tbl.PK), KeyType: types.KeyTypeHash},
	}

	var gsiInputs []types.GlobalSecondaryIndex
	for _, gsi := range tbl.GSIs {
		gsiKeySchema := []types.KeySchemaElement{
			{AttributeName: aws.String(gsi.PK), KeyType: types.KeyTypeHash},
		}
		if gsi.SK != "" {
			gsiKeySchema = append(gsiKeySchema, types.KeySchemaElement{
				AttributeName: aws.String(gsi.SK), KeyType: types.KeyTypeRange,
			})
		}

		gsiInputs = append(gsiInputs, types.GlobalSecondaryIndex{
			IndexName: aws.String(gsi.Name),
			KeySchema: gsiKeySchema,
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeAll,
			},
		})
	}

	input := &dynamodb.CreateTableInput{
		TableName:            aws.String(tbl.Name),
		AttributeDefinitions: attrDefs,
		KeySchema:            keySchema,
		GlobalSecondaryIndexes: gsiInputs,
		BillingMode:          types.BillingModePayPerRequest,
	}

	_, err := svc.CreateTable(ctx, input)
	if err != nil {
		return err
	}

	// Wait for table to become active
	waiter := dynamodb.NewTableExistsWaiter(svc)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tbl.Name),
	}, 30*time.Second)
}

func seedTodaysQuestion(ctx context.Context, client *db.DynamoDBClient) error {
	today := time.Now().UTC().Format("2006-01-02")
	existing, err := client.GetQuestion(ctx, today)
	if err != nil {
		return fmt.Errorf("check existing question: %w", err)
	}
	if existing != nil {
		return nil
	}

	question := &models.Question{
		Date:        today,
		QuestionID:  uuid.New().String(),
		Title:       "Design a URL Shortener",
		Description: "Design a URL shortening service like TinyURL. The system should take a long URL and generate a short, unique alias for it. Users should be able to use the short URL to redirect to the original URL.",
		Difficulty:  "Medium",
		Categories:  []string{"Scalability", "Database Design", "API Design"},
		Hints:       []string{"Think about hashing vs base62 encoding", "Consider read vs write ratio", "How would you handle collisions?"},
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := client.PutQuestion(ctx, question); err != nil {
		return fmt.Errorf("seed question: %w", err)
	}
	log.Printf("Seeded question for %s", today)
	return nil
}

func handleGetQuestion(client *db.DynamoDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			date = time.Now().UTC().Format("2006-01-02")
		}

		question, err := client.GetQuestion(r.Context(), date)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if question == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "No question for this date"})
			return
		}

		writeJSON(w, http.StatusOK, question)
	}
}

func handleSubmitAnswer(client *db.DynamoDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SubmitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			return
		}

		if req.UserID == "" || req.QuestionDate == "" || req.Answer == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "userId, questionDate, and answer are required"})
			return
		}

		submissionID := uuid.New().String()
		now := time.Now().UTC().Format(time.RFC3339)

		submission := &models.Submission{
			SubmissionID: submissionID,
			UserID:       req.UserID,
			QuestionDate: req.QuestionDate,
			Answer:       req.Answer,
			Status:       "pending",
			SubmittedAt:  now,
		}

		if err := client.PutSubmission(r.Context(), submission); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		submissionEvals.Store(submissionID, submission)

		evalWorkers.Add(1)
		go evaluateSubmission(client, submissionID)

		writeJSON(w, http.StatusAccepted, map[string]string{
			"submissionId": submissionID,
			"status":       "pending",
		})
	}
}

func evaluateSubmission(client *db.DynamoDBClient, submissionID string) {
	defer evalWorkers.Done()
	time.Sleep(3 * time.Second)

	categories := map[string]models.CategoryScore{
		"Requirements": {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Requirements")},
		"API Design":   {Score: rand.Intn(41) + 60, Feedback: randomFeedback("API Design")},
		"Database":     {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Database")},
		"Cache":        {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Cache")},
		"Scalability":  {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Scalability")},
		"Availability": {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Availability")},
		"Tradeoffs":    {Score: rand.Intn(41) + 60, Feedback: randomFeedback("Tradeoffs")},
	}

	total := 0
	for _, c := range categories {
		total += c.Score
	}
	overallScore := total / len(categories)

	strengths := []string{
		"Good understanding of distributed systems concepts",
		"Clear and structured approach to the problem",
		"Well-considered trade-offs in the design",
	}
	weaknesses := []string{
		"Could add more detail on data consistency mechanisms",
		"Consider discussing monitoring and observability",
	}

	summary := "Overall, this is a solid solution. The candidate demonstrates good system design thinking with clear trade-off analysis."
	if overallScore < 70 {
		summary = "The solution covers the basics but needs more depth in distributed systems considerations and edge cases."
	} else if overallScore > 85 {
		strengths = append(strengths, "Excellent depth in distributed systems knowledge")
		summary = "Outstanding solution with comprehensive coverage of system design aspects."
	}

	val, ok := submissionEvals.Load(submissionID)
	if !ok {
		return
	}
	sub := val.(*models.Submission)
	sub.Status = "evaluated"
	sub.Score = overallScore
	sub.Evaluation = &models.EvaluationResult{
		Score:      overallScore,
		Strengths:  strengths,
		Weaknesses: weaknesses,
		Categories: categories,
		Summary:    summary,
	}
	sub.EvaluatedAt = time.Now().UTC().Format(time.RFC3339)
	submissionEvals.Store(submissionID, sub)

	if err := client.UpdateSubmissionResult(context.Background(), submissionID, sub.Evaluation, overallScore); err != nil {
		log.Printf("Failed to persist evaluation to DynamoDB: %v", err)
	}

	question, _ := client.GetQuestion(context.Background(), sub.QuestionDate)
	questionTitle := "Architecture Challenge"
	if question != nil {
		questionTitle = question.Title
	}
	if err := client.UpdateLeaderboard(context.Background(), sub.QuestionDate, overallScore, sub.UserID, questionTitle); err != nil {
		log.Printf("Failed to update leaderboard in DynamoDB: %v", err)
	}

	log.Printf("Evaluated submission %s: score=%d", submissionID, overallScore)
}

func handleGetResults(client *db.DynamoDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		submissionID := r.PathValue("submissionId")
		if submissionID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "submissionId is required"})
			return
		}

		val, ok := submissionEvals.Load(submissionID)
		if ok {
			writeJSON(w, http.StatusOK, val.(*models.Submission))
			return
		}

		sub, err := client.GetSubmission(r.Context(), submissionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Submission not found"})
			return
		}

		writeJSON(w, http.StatusOK, sub)
	}
}

func handleGetLeaderboard(client *db.DynamoDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			date = time.Now().UTC().Format("2006-01-02")
		}

		questionTitle := "Architecture Challenge"
		question, err := client.GetQuestion(r.Context(), date)
		if err == nil && question != nil {
			questionTitle = question.Title
		}

		var topSubmissions []models.Submission
		submissionEvals.Range(func(key, val any) bool {
			sub := val.(*models.Submission)
			if sub.QuestionDate == date && sub.Status == "evaluated" {
				topSubmissions = append(topSubmissions, *sub)
			}
			return true
		})

		dbSubs, err := client.QueryTopSubmissions(r.Context(), date, 100)
		if err == nil {
			seen := map[string]bool{}
			for _, s := range topSubmissions {
				seen[s.SubmissionID] = true
			}
			for _, s := range dbSubs {
				if !seen[s.SubmissionID] {
					topSubmissions = append(topSubmissions, s)
				}
			}
		}

		sort.Slice(topSubmissions, func(i, j int) bool {
			return topSubmissions[i].Score > topSubmissions[j].Score
		})

		if topSubmissions == nil {
			topSubmissions = []models.Submission{}
		}

		highestScore := 0
		topScorer := "N/A"
		if len(topSubmissions) > 0 {
			highestScore = topSubmissions[0].Score
			topScorer = topSubmissions[0].UserID
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"leaderboard": map[string]any{
				"date":             date,
				"highestScore":     highestScore,
				"topScorer":        topScorer,
				"totalSubmissions": len(topSubmissions),
				"questionTitle":    questionTitle,
			},
			"topSubmissions": topSubmissions,
		})
	}
}

func handleGetUserStats(client *db.DynamoDBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userId")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "userId is required"})
			return
		}

		var submissions []*models.Submission
		seen := map[string]bool{}
		submissionEvals.Range(func(key, val any) bool {
			sub := val.(*models.Submission)
			if sub.UserID == userID {
				submissions = append(submissions, sub)
				seen[sub.SubmissionID] = true
			}
			return true
		})

		dbSubs, err := client.QuerySubmissionsByUser(r.Context(), userID)
		if err == nil {
			for i := range dbSubs {
				if !seen[dbSubs[i].SubmissionID] {
					submissions = append(submissions, &dbSubs[i])
				}
			}
		}

		totalScore := 0
		evaluatedCount := 0
		for _, s := range submissions {
			if s.Status == "evaluated" {
				totalScore += s.Score
				evaluatedCount++
			}
		}

		avgScore := 0.0
		if evaluatedCount > 0 {
			avgScore = float64(totalScore) / float64(evaluatedCount)
		}

		sort.Slice(submissions, func(i, j int) bool {
			return submissions[i].SubmittedAt > submissions[j].SubmittedAt
		})

		longestStreak := 0
		currentStreak := 0
		if len(submissions) > 0 {
			longestStreak = 1
			currentStreak = 1
			streak := 1
			for i := 1; i < len(submissions); i++ {
				prevDate, _ := time.Parse(time.RFC3339, submissions[i-1].SubmittedAt)
				currDate, _ := time.Parse(time.RFC3339, submissions[i].SubmittedAt)
				diff := prevDate.Sub(currDate).Hours()
				if diff <= 48 {
					streak++
					if streak > longestStreak {
						longestStreak = streak
					}
				} else {
					streak = 1
				}
			}

			if time.Since(time.Now().Add(-48*time.Hour)) != time.Duration(0) {
				currentStreak = 1
				for i := 0; i < len(submissions)-1; i++ {
					d1, _ := time.Parse(time.RFC3339, submissions[i].SubmittedAt)
					d2, _ := time.Parse(time.RFC3339, submissions[i+1].SubmittedAt)
					if d1.Sub(d2).Hours() <= 48 {
						currentStreak++
					} else {
						break
					}
				}
			}
		}

		recent := make([]models.SubmissionSummary, 0, len(submissions))
		for _, s := range submissions {
			recent = append(recent, models.SubmissionSummary{
				SubmissionID: s.SubmissionID,
				QuestionDate: s.QuestionDate,
				Score:        s.Score,
				Status:       s.Status,
				SubmittedAt:  s.SubmittedAt,
			})
		}
		if len(recent) > 20 {
			recent = recent[:20]
		}

		writeJSON(w, http.StatusOK, models.UserStats{
			UserID:            userID,
			TotalSubmissions:  len(submissions),
			AverageScore:      avgScore,
			CurrentStreak:     currentStreak,
			LongestStreak:     longestStreak,
			WeakAreas:         []models.WeakArea{},
			RecentSubmissions: recent,
		})
	}
}

func toAttrType(t string) types.ScalarAttributeType {
	if t == "N" {
		return types.ScalarAttributeTypeN
	}
	return types.ScalarAttributeTypeS
}

func randomFeedback(category string) string {
	feedbacks := []string{
		"Good consideration of " + category + " aspects",
		"Adequate coverage of " + category + " principles",
		"Room for improvement in " + category + " depth",
		"Well-structured approach to " + category,
		"Solid understanding of " + category + " concepts",
	}
	return feedbacks[rand.Intn(len(feedbacks))]
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

const swaggerUIPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>AI Interviewer API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body style="margin:0">
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({ url: "/swagger.json", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`
