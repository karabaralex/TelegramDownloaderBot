package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var API_KEY string

// GenerateAiResponse uses the GenAI client to generate content based on the given text.
// It requires an API key and the text for content generation as inputs.
func GenerateAiResponse(text string) (string, error) {
	ctx := context.Background()

	// Set up the client with the provided API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(API_KEY))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return "", err
	}
	defer client.Close()

	// Specify the model for text-only input
	model := client.GenerativeModel("gemini-pro")

	// Generate content based on the provided text
	resp, err := model.GenerateContent(ctx, genai.Text(text))
	if resp == nil && err != nil {
		log.Fatalf("Failed to generate content: %v", err)
		return "", err
	}

	return stringResponse(resp), nil
}

func stringResponse(resp *genai.GenerateContentResponse) string {
	result := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				result += fmt.Sprintln(part)
			}
		}
	}

	return result
}
