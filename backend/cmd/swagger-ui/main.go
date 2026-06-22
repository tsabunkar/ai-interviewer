package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "GET,OPTIONS",
	}

	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	if req.Resource == "/swagger.json" || strings.HasSuffix(req.Path, "/swagger.json") {
		headers["Content-Type"] = "application/json"
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       getSwaggerSpec(),
		}, nil
	}

	headers["Content-Type"] = "text/html; charset=utf-8"
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       swaggerUIPage,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}

func getSwaggerSpec() string {
	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}

	parsed, err := url.Parse(apiBaseURL)
	if err != nil {
		parsed, _ = url.Parse("http://localhost:8080")
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "http"
	}

	host := parsed.Host
	basePath := parsed.Path
	if basePath == "" || basePath == "/" {
		basePath = ""
	}

	replacer := strings.NewReplacer(
		"__HOST__", host,
		"__SCHEME__", scheme,
		"__BASE_PATH__", basePath,
	)

	return replacer.Replace(swaggerSpecTemplate)
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
    var baseUrl = window.location.pathname.replace(/\/?$/, '');
    SwaggerUIBundle({ url: baseUrl + '/swagger.json', dom_id: "#swagger-ui" });
  </script>
</body>
</html>`

const swaggerSpecTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "Backend services for the AI Interviewer application",
    "title": "AI Interviewer API",
    "version": "1.0.0"
  },
  "host": "__HOST__",
  "schemes": ["__SCHEME__"],
  "basePath": "__BASE_PATH__",
  "paths": {
    "/question": {
      "get": {
        "summary": "Fetch daily coding question",
        "description": "Retrieves the coding question for the current date",
        "operationId": "fetchQuestion",
        "responses": {
          "200": { "description": "Successfully retrieved the question" },
          "404": { "description": "No question available for today" },
          "500": { "description": "Internal server error" }
        }
      }
    },
    "/answer": {
      "post": {
        "summary": "Submit coding answer",
        "description": "Submit a coding answer for evaluation",
        "operationId": "submitAnswer",
        "parameters": [{
          "name": "body",
          "in": "body",
          "required": true,
          "schema": { "type": "object" }
        }],
        "responses": {
          "202": { "description": "Answer accepted for processing" },
          "400": { "description": "Invalid request payload" },
          "500": { "description": "Internal server error" }
        }
      }
    },
    "/results/{submissionId}": {
      "get": {
        "summary": "Get evaluation results",
        "description": "Retrieve the results of answer evaluations",
        "operationId": "getResults",
        "parameters": [{
          "name": "submissionId",
          "in": "path",
          "required": true,
          "type": "string"
        }],
        "responses": {
          "200": { "description": "Successfully retrieved results" },
          "500": { "description": "Internal server error" }
        }
      }
    },
    "/leaderboard": {
      "get": {
        "summary": "Get leaderboard rankings",
        "description": "Retrieve user rankings based on performance",
        "operationId": "getLeaderboard",
        "responses": {
          "200": { "description": "Successfully retrieved leaderboard" },
          "500": { "description": "Internal server error" }
        }
      }
    },
    "/stats/{userId}": {
      "get": {
        "summary": "Get user statistics",
        "description": "Retrieve statistics for a specific user",
        "operationId": "getUserStats",
        "parameters": [{
          "name": "userId",
          "in": "path",
          "required": true,
          "type": "string"
        }],
        "responses": {
          "200": { "description": "Successfully retrieved user stats" },
          "500": { "description": "Internal server error" }
        }
      }
    }
  }
}`
