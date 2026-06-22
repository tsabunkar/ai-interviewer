package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/go-openapi/spec"
)

func main() {
	// Create the base OpenAPI specification
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       "AI Interviewer API",
					Description: "Backend services for the AI Interviewer application",
					Version:     "1.0.0",
				},
			},
			Schemes: []string{"https"},
			Paths:   &spec.Paths{},
		},
	}

	// Add paths for each Lambda function
	paths := &spec.Paths{
		Paths: map[string]spec.PathItem{
			"/question": {
				PathItemProps: spec.PathItemProps{
					Get: &spec.Operation{
						OperationProps: spec.OperationProps{
							Summary:     "Fetch daily coding question",
							Description: "Retrieves the coding question for the current date",
							ID:          "fetchQuestion",
							Responses: &spec.Responses{
								ResponsesProps: spec.ResponsesProps{
									StatusCodeResponses: map[int]spec.Response{
										200: {
											ResponseProps: spec.ResponseProps{
												Description: "Successfully retrieved the question",
											},
										},
										404: {
											ResponseProps: spec.ResponseProps{
												Description: "No question available for today",
											},
										},
										500: {
											ResponseProps: spec.ResponseProps{
												Description: "Internal server error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/answer": {
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						OperationProps: spec.OperationProps{
							Summary:     "Submit coding answer",
							Description: "Submit a coding answer for evaluation",
							ID:          "submitAnswer",
							Parameters: []spec.Parameter{
								{
									ParamProps: spec.ParamProps{
										Name:     "body",
										In:       "body",
										Required: true,
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type: spec.StringOrArray{"object"},
											},
										},
									},
								},
							},
							Responses: &spec.Responses{
								ResponsesProps: spec.ResponsesProps{
									StatusCodeResponses: map[int]spec.Response{
										202: {
											ResponseProps: spec.ResponseProps{
												Description: "Answer accepted for processing",
											},
										},
										400: {
											ResponseProps: spec.ResponseProps{
												Description: "Invalid request payload",
											},
										},
										500: {
											ResponseProps: spec.ResponseProps{
												Description: "Internal server error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/results": {
				PathItemProps: spec.PathItemProps{
					Get: &spec.Operation{
						OperationProps: spec.OperationProps{
							Summary:     "Get evaluation results",
							Description: "Retrieve the results of answer evaluations",
							ID:          "getResults",
							Responses: &spec.Responses{
								ResponsesProps: spec.ResponsesProps{
									StatusCodeResponses: map[int]spec.Response{
										200: {
											ResponseProps: spec.ResponseProps{
												Description: "Successfully retrieved results",
											},
										},
										500: {
											ResponseProps: spec.ResponseProps{
												Description: "Internal server error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/leaderboard": {
				PathItemProps: spec.PathItemProps{
					Get: &spec.Operation{
						OperationProps: spec.OperationProps{
							Summary:     "Get leaderboard rankings",
							Description: "Retrieve user rankings based on performance",
							ID:          "getLeaderboard",
							Responses: &spec.Responses{
								ResponsesProps: spec.ResponsesProps{
									StatusCodeResponses: map[int]spec.Response{
										200: {
											ResponseProps: spec.ResponseProps{
												Description: "Successfully retrieved leaderboard",
											},
										},
										500: {
											ResponseProps: spec.ResponseProps{
												Description: "Internal server error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/stats": {
				PathItemProps: spec.PathItemProps{
					Get: &spec.Operation{
						OperationProps: spec.OperationProps{
							Summary:     "Get user statistics",
							Description: "Retrieve statistics for a specific user",
							ID:          "getUserStats",
							Responses: &spec.Responses{
								ResponsesProps: spec.ResponsesProps{
									StatusCodeResponses: map[int]spec.Response{
										200: {
											ResponseProps: spec.ResponseProps{
												Description: "Successfully retrieved user stats",
											},
										},
										500: {
											ResponseProps: spec.ResponseProps{
												Description: "Internal server error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	swagger.Paths = paths

	// Convert to JSON and write to file
	data, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("docs/swagger.json", data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Swagger documentation generated successfully!")
}