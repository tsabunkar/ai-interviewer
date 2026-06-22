package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ai-interviewer/backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockClient struct {
	Client  *bedrockruntime.Client
	ModelID string
}

type BedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	Temperature      float64          `json:"temperature,omitempty"`
	System           string           `json:"system,omitempty"`
	Messages         []BedrockMessage `json:"messages"`
}

type BedrockMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type BedrockResponse struct {
	Content []ContentBlock `json:"content"`
}

func NewBedrockClient(ctx context.Context) (*BedrockClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load SDK config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "anthropic.claude-3-sonnet-20240229-v1:0"
	}

	return &BedrockClient{
		Client:  client,
		ModelID: modelID,
	}, nil
}

func (b *BedrockClient) invokeClaude(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4000,
		Temperature:      0.7,
		System:           systemPrompt,
		Messages: []BedrockMessage{
			{
				Role: "user",
				Content: []ContentBlock{
					{Type: "text", Text: userMessage},
				},
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Bedrock request: %w", err)
	}

	output, err := b.Client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(b.ModelID),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke Bedrock model %s: %w", b.ModelID, err)
	}

	var response BedrockResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	if len(response.Content) > 0 {
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("empty response content from Bedrock")
}

func extractJSON(response string) (string, error) {
	// Find the JSON boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start > end {
		return "", fmt.Errorf("could not find JSON block in response")
	}
	return response[start : end+1], nil
}

func (b *BedrockClient) GenerateQuestion(ctx context.Context) (*models.Question, error) {
	systemPrompt := "You are a master system design interviewer. Generate a challenging system design question for a daily architecture competition. Return ONLY a valid JSON object. Do not include markdown codeblocks or extra conversational text."
	userMessage := `Generate a system design question. The output MUST be a JSON object with this exact structure:
{
  "title": "A short, catchy title of the system to design (e.g. Design a Distributed Rate Limiter)",
  "description": "A detailed problem statement, functional requirements, non-functional requirements, constraints, traffic estimates, and expectations. Use markdown format inside this string. Make it extensive, realistic, and clear.",
  "difficulty": "Easy, Medium, or Hard",
  "categories": ["Requirements", "API Design", "Database", "Cache", "Scalability", "Availability", "Tradeoffs"],
  "hints": ["Hint 1", "Hint 2", "Hint 3"]
}`

	respText, err := b.invokeClaude(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	jsonStr, err := extractJSON(respText)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from response: %w. Raw response: %s", err, respText)
	}

	var question models.Question
	if err := json.Unmarshal([]byte(jsonStr), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question JSON: %w. JSON: %s", err, jsonStr)
	}

	return &question, nil
}

func (b *BedrockClient) EvaluateAnswer(ctx context.Context, questionDesc, answer string) (*models.EvaluationResult, error) {
	systemPrompt := "You are a senior principal systems architect conducting a system design interview. You evaluate candidate answers deeply and return ONLY a valid JSON object. Do not include markdown codeblocks or extra conversational text."
	
	userMessage := fmt.Sprintf(`Evaluate the candidate's answer for the following System Design Question.

### Question
%s

### Candidate Answer
%s

You must rate their answer across the following 7 categories:
1. Requirements (Functional and Non-Functional requirement gathering and clarity)
2. API Design (Endpoints, request/response structures, protocol choice)
3. Database (Schema, choices of DB, data partitioning, indexing)
4. Cache (Caching strategy, eviction, cache invalidation, cache locations)
5. Scalability (Load balancing, horizontal scaling, data processing pipelines, messaging)
6. Availability (Replication, failover, disaster recovery, rate limiting, fault tolerance)
7. Tradeoffs (Deep analysis of choices, bottlenecks, alternatives, CAP theorem, costing)

For each category, provide a score (0 to 100) and constructive feedback.
Calculate an overall weighted score (0 to 100).
Highlight specific strengths (array of strings) and weaknesses (array of strings).
Provide an overall summary feedback (string).

Return a JSON object matching this exact schema:
{
  "score": 75,
  "strengths": [
    "Strength 1",
    "Strength 2"
  ],
  "weaknesses": [
    "Weakness 1",
    "Weakness 2"
  ],
  "categories": {
    "Requirements": {
      "score": 80,
      "feedback": "Feedback for Requirements"
    },
    "API Design": {
      "score": 70,
      "feedback": "Feedback for API Design"
    },
    "Database": {
      "score": 60,
      "feedback": "Feedback for Database"
    },
    "Cache": {
      "score": 65,
      "feedback": "Feedback for Cache"
    },
    "Scalability": {
      "score": 85,
      "feedback": "Feedback for Scalability"
    },
    "Availability": {
      "score": 80,
      "feedback": "Feedback for Availability"
    },
    "Tradeoffs": {
      "score": 70,
      "feedback": "Feedback for Tradeoffs"
    }
  },
  "summary": "Overall summary of the candidate's submission."
}`, questionDesc, answer)

	respText, err := b.invokeClaude(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	jsonStr, err := extractJSON(respText)
	if err != nil {
		return nil, fmt.Errorf("failed to extract evaluation JSON from response: %w. Raw response: %s", err, respText)
	}

	var result models.EvaluationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse evaluation JSON: %w. JSON: %s", err, jsonStr)
	}

	return &result, nil
}
