package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/internal/config"
)

const (
	defaultIdeaCount     = 3
	maxIdeaCount         = 5
	yandexCompletionURL  = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
	yandexIdeaSourceName = "llm:yandexgpt"
)

var (
	ErrLLMUnavailable      = errors.New("idea generation service is not configured")
	ErrIdeaTopicRequired   = errors.New("topic is required")
	ErrIdeaCountOutOfRange = errors.New("count must be between 1 and 5")
	ErrLLMInvalidResponse  = errors.New("idea generation returned invalid response")
)

type IdeaGenerator interface {
	GenerateIdeas(ctx context.Context, req IdeaGenerationRequest) ([]GeneratedIdeaContent, error)
}

type IdeaGenerationRequest struct {
	CompanyName string
	Topic       string
	Context     *string
	Audience    *string
	Constraints []string
	Tone        *string
	Count       int
}

type GeneratedIdeaContent struct {
	Title       string
	Description string
}

type disabledIdeaGenerator struct{}

func (disabledIdeaGenerator) GenerateIdeas(context.Context, IdeaGenerationRequest) ([]GeneratedIdeaContent, error) {
	return nil, ErrLLMUnavailable
}

type YandexIdeaGenerator struct {
	apiKey   string
	folderID string
	model    string
	client   *http.Client
}

func NewIdeaGenerator(cfg config.Config) IdeaGenerator {
	if strings.TrimSpace(cfg.LLMProvider) == "" {
		return disabledIdeaGenerator{}
	}

	switch strings.ToLower(strings.TrimSpace(cfg.LLMProvider)) {
	case "yandex":
		if strings.TrimSpace(cfg.YandexAPIKey) == "" || strings.TrimSpace(cfg.YandexFolderID) == "" {
			return disabledIdeaGenerator{}
		}
		timeout := time.Duration(cfg.LLMTimeoutSec) * time.Second
		if timeout <= 0 {
			timeout = 15 * time.Second
		}
		return &YandexIdeaGenerator{
			apiKey:   cfg.YandexAPIKey,
			folderID: cfg.YandexFolderID,
			model:    cfg.YandexModel,
			client:   &http.Client{Timeout: timeout},
		}
	default:
		return disabledIdeaGenerator{}
	}
}

func (g *YandexIdeaGenerator) GenerateIdeas(ctx context.Context, req IdeaGenerationRequest) ([]GeneratedIdeaContent, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, ErrIdeaTopicRequired
	}

	count := req.Count
	if count == 0 {
		count = defaultIdeaCount
	}
	if count < 1 || count > maxIdeaCount {
		return nil, ErrIdeaCountOutOfRange
	}

	prompt := buildIdeaPrompt(req, count)
	requestBody := yandexCompletionRequest{
		ModelURI: fmt.Sprintf("gpt://%s/%s", g.folderID, strings.TrimSpace(g.model)),
		CompletionOptions: yandexCompletionOptions{
			Stream:      false,
			Temperature: 0.8,
			MaxTokens:   "1800",
		},
		Messages: []yandexMessage{
			{
				Role: "system",
				Text: "You generate concise and original ideas for meetings. " +
					"Always follow the required JSON schema exactly. " +
					"Answer in the user's language.",
			},
			{
				Role: "user",
				Text: prompt,
			},
		},
		JSONSchema: yandexJSONSchema{
			Name:        "generated_ideas",
			Description: "A list of generated idea drafts",
			Schema:      ideaSchema(count),
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexCompletionURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Api-Key "+g.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("idea generation request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("idea generation request failed: %s", strings.TrimSpace(string(respBody)))
	}

	var completionResp yandexCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLLMInvalidResponse, err)
	}
	if len(completionResp.Result.Alternatives) == 0 {
		return nil, ErrLLMInvalidResponse
	}

	var parsed struct {
		Ideas []GeneratedIdeaContent `json:"ideas"`
	}
	if err := json.Unmarshal([]byte(completionResp.Result.Alternatives[0].Message.Text), &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLLMInvalidResponse, err)
	}
	if len(parsed.Ideas) == 0 {
		return nil, ErrLLMInvalidResponse
	}

	result := make([]GeneratedIdeaContent, 0, len(parsed.Ideas))
	for _, idea := range parsed.Ideas {
		title := strings.TrimSpace(idea.Title)
		description := strings.TrimSpace(idea.Description)
		if title == "" || description == "" {
			return nil, ErrLLMInvalidResponse
		}
		result = append(result, GeneratedIdeaContent{
			Title:       title,
			Description: description,
		})
	}

	return result, nil
}

func buildIdeaPrompt(req IdeaGenerationRequest, count int) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Company name: %s.", req.CompanyName))
	parts = append(parts, fmt.Sprintf("Main topic: %s.", strings.TrimSpace(req.Topic)))
	if req.Context != nil && strings.TrimSpace(*req.Context) != "" {
		parts = append(parts, fmt.Sprintf("Context: %s.", strings.TrimSpace(*req.Context)))
	}
	if req.Audience != nil && strings.TrimSpace(*req.Audience) != "" {
		parts = append(parts, fmt.Sprintf("Target audience: %s.", strings.TrimSpace(*req.Audience)))
	}
	if req.Tone != nil && strings.TrimSpace(*req.Tone) != "" {
		parts = append(parts, fmt.Sprintf("Tone: %s.", strings.TrimSpace(*req.Tone)))
	}
	if len(req.Constraints) > 0 {
		cleanConstraints := make([]string, 0, len(req.Constraints))
		for _, constraint := range req.Constraints {
			constraint = strings.TrimSpace(constraint)
			if constraint != "" {
				cleanConstraints = append(cleanConstraints, constraint)
			}
		}
		if len(cleanConstraints) > 0 {
			parts = append(parts, "Constraints: "+strings.Join(cleanConstraints, "; ")+".")
		}
	}
	parts = append(parts, fmt.Sprintf("Generate exactly %d distinct ideas.", count))
	parts = append(parts, "Each idea must have a short title and a description of 2-4 sentences with concrete value.")
	parts = append(parts, "Do not include markdown, numbering, comments, or fields outside the schema.")
	return strings.Join(parts, " ")
}

func ideaSchema(count int) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ideas": map[string]any{
				"type":     "array",
				"minItems": count,
				"maxItems": count,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "Short idea title",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Detailed description of the idea",
						},
					},
					"required": []string{
						"title",
						"description",
					},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"ideas"},
		"additionalProperties": false,
	}
}

type yandexCompletionRequest struct {
	ModelURI          string                  `json:"modelUri"`
	CompletionOptions yandexCompletionOptions `json:"completionOptions"`
	Messages          []yandexMessage         `json:"messages"`
	JSONSchema        yandexJSONSchema        `json:"jsonSchema"`
}

type yandexCompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   string  `json:"maxTokens"`
}

type yandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type yandexJSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
}

type yandexCompletionResponse struct {
	Result struct {
		Alternatives []struct {
			Message yandexMessage `json:"message"`
		} `json:"alternatives"`
	} `json:"result"`
}
