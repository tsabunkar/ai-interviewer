package models

// Question represents a daily system design interview question.
type Question struct {
	Date        string   `json:"date" dynamodbav:"date"`
	QuestionID  string   `json:"questionId" dynamodbav:"questionId"`
	Title       string   `json:"title" dynamodbav:"title"`
	Description string   `json:"description" dynamodbav:"description"`
	Difficulty  string   `json:"difficulty" dynamodbav:"difficulty"`
	Categories  []string `json:"categories" dynamodbav:"categories"`
	Hints       []string `json:"hints" dynamodbav:"hints"`
	CreatedAt   string   `json:"createdAt" dynamodbav:"createdAt"`
}

// Submission represents a user's answer submission and its evaluation.
type Submission struct {
	SubmissionID string            `json:"submissionId" dynamodbav:"submissionId"`
	UserID       string            `json:"userId" dynamodbav:"userId"`
	QuestionDate string            `json:"questionDate" dynamodbav:"questionDate"`
	Answer       string            `json:"answer" dynamodbav:"answer"`
	Status       string            `json:"status" dynamodbav:"status"` // pending|evaluated
	Score        int               `json:"score" dynamodbav:"score"`
	Evaluation   *EvaluationResult `json:"evaluation,omitempty" dynamodbav:"evaluation,omitempty"`
	SubmittedAt  string            `json:"submittedAt" dynamodbav:"submittedAt"`
	EvaluatedAt  string            `json:"evaluatedAt,omitempty" dynamodbav:"evaluatedAt,omitempty"`
}

// EvaluationResult contains the AI-generated evaluation of a submission.
type EvaluationResult struct {
	Score      int                      `json:"score"`
	Strengths  []string                 `json:"strengths"`
	Weaknesses []string                 `json:"weaknesses"`
	Categories map[string]CategoryScore `json:"categories"`
	Summary    string                   `json:"summary"`
}

// CategoryScore holds the score and feedback for a specific evaluation category.
type CategoryScore struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// LeaderboardEntry represents a single day's leaderboard state.
type LeaderboardEntry struct {
	Date             string `json:"date" dynamodbav:"date"`
	HighestScore     int    `json:"highestScore" dynamodbav:"highestScore"`
	TopScorer        string `json:"topScorer" dynamodbav:"topScorer"`
	TotalSubmissions int    `json:"totalSubmissions" dynamodbav:"totalSubmissions"`
	QuestionTitle    string `json:"questionTitle" dynamodbav:"questionTitle"`
}

// SubmitRequest is the API request body for submitting an answer.
type SubmitRequest struct {
	UserID       string `json:"userId"`
	QuestionDate string `json:"questionDate"`
	Answer       string `json:"answer"`
}

// SQSSubmissionMessage is the message format sent to the evaluation SQS queue.
type SQSSubmissionMessage struct {
	SubmissionID string `json:"submissionId"`
	UserID       string `json:"userId"`
	QuestionDate string `json:"questionDate"`
	Answer       string `json:"answer"`
}

// UserStats contains aggregated statistics for a user.
type UserStats struct {
	UserID            string              `json:"userId"`
	TotalSubmissions  int                 `json:"totalSubmissions"`
	AverageScore      float64             `json:"averageScore"`
	CurrentStreak     int                 `json:"currentStreak"`
	LongestStreak     int                 `json:"longestStreak"`
	WeakAreas         []WeakArea          `json:"weakAreas"`
	RecentSubmissions []SubmissionSummary `json:"recentSubmissions"`
}

// WeakArea identifies a category where the user scores below average.
type WeakArea struct {
	Category     string  `json:"category"`
	AverageScore float64 `json:"averageScore"`
}

// SubmissionSummary is a condensed view of a submission for the stats endpoint.
type SubmissionSummary struct {
	SubmissionID string `json:"submissionId"`
	QuestionDate string `json:"questionDate"`
	Score        int    `json:"score"`
	Status       string `json:"status"`
	SubmittedAt  string `json:"submittedAt"`
}
