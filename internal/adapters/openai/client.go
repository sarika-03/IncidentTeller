package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"incident-teller/internal/config"
	"incident-teller/internal/domain"
)

// Client provides OpenAI API integration
type Client struct {
	apiClient *openai.Client
	config    config.OpenAIConfig
}

// NewClient creates a new OpenAI client
func NewClient(cfg config.OpenAIConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("OpenAI is not enabled in configuration")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	apiClient := openai.NewClient(cfg.APIKey)

	return &Client{
		apiClient: apiClient,
		config:    cfg,
	}, nil
}

// AnalyzeIncident generates a summary and insights about an incident
func (c *Client) AnalyzeIncident(ctx context.Context, alerts []domain.Alert) (IncidentAnalysis, error) {
	if len(alerts) == 0 {
		return IncidentAnalysis{}, fmt.Errorf("no alerts to analyze")
	}

	// Prepare context from alerts
	context := c.prepareIncidentContext(alerts)

	// Generate summary
	summary, err := c.generateIncidentSummary(ctx, context)
	if err != nil {
		return IncidentAnalysis{}, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Generate root cause analysis
	rootCause, err := c.generateRootCauseAnalysis(ctx, context)
	if err != nil {
		return IncidentAnalysis{}, fmt.Errorf("failed to generate root cause analysis: %w", err)
	}

	// Generate recommendations
	recommendations, err := c.generateRecommendations(ctx, context, rootCause)
	if err != nil {
		return IncidentAnalysis{}, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Generate impact assessment
	impact, err := c.generateImpactAssessment(ctx, context)
	if err != nil {
		return IncidentAnalysis{}, fmt.Errorf("failed to generate impact assessment: %w", err)
	}

	return IncidentAnalysis{
		Summary:         summary,
		RootCause:       rootCause,
		Recommendations: recommendations,
		Impact:          impact,
		GeneratedAt:     time.Now(),
		AlertCount:      len(alerts),
		TimeSpan:        c.calculateTimeSpan(alerts),
	}, nil
}

// generateIncidentSummary creates a concise summary of the incident
func (c *Client) generateIncidentSummary(ctx context.Context, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert SRE analyzing a system incident. Based on the following incident data, provide a concise summary (2-3 sentences) of what happened.

Incident Data:
%s

Summary:`, context)

	response, err := c.callOpenAI(ctx, prompt, "Provide a clear, concise summary of the incident.")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// generateRootCauseAnalysis identifies the root cause of the incident
func (c *Client) generateRootCauseAnalysis(ctx context.Context, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert SRE analyzing a system incident. Based on the following incident data, identify the most likely root cause. Consider cascading failures, resource exhaustion, and system interactions.

Incident Data:
%s

Root Cause Analysis:
1. Primary cause:
2. Contributing factors:
3. Why it occurred:`, context)

	response, err := c.callOpenAI(ctx, prompt, "Identify the root cause of the incident.")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// generateRecommendations provides actionable recommendations
func (c *Client) generateRecommendations(ctx context.Context, context string, rootCause string) (Recommendations, error) {
	prompt := fmt.Sprintf(`You are an expert SRE analyzing a system incident. Based on the incident and root cause analysis, provide actionable recommendations organized by timeframe.

Incident Data:
%s

Root Cause:
%s

Provide recommendations in this format:
IMMEDIATE (within 5 minutes):
- Action 1
- Action 2

SHORT TERM (within 8 hours):
- Action 1
- Action 2

LONG TERM (prevention, ongoing):
- Action 1
- Action 2`, context, rootCause)

	response, err := c.callOpenAI(ctx, prompt, "Generate actionable recommendations.")
	if err != nil {
		return Recommendations{}, err
	}

	return c.parseRecommendations(response), nil
}

// generateImpactAssessment assesses the business impact
func (c *Client) generateImpactAssessment(ctx context.Context, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert SRE analyzing a system incident. Based on the incident data, provide a brief assessment of the business impact (1-2 sentences).

Incident Data:
%s

Consider:
- Services affected
- Severity (low/medium/high/critical)
- User impact
- Data integrity concerns

Impact Assessment:`, context)

	response, err := c.callOpenAI(ctx, prompt, "Assess the business impact.")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// callOpenAI makes a request to the OpenAI API
func (c *Client) callOpenAI(ctx context.Context, prompt string, system string) (string, error) {
	// Create a timeout context if one doesn't exist
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()
	}

	req := openai.ChatCompletionRequest{
		Model: c.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: float32(c.config.Temperature),
		TopP:        float32(c.config.TopP),
	}

	resp, err := c.apiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return resp.Choices[0].Message.Content, nil
}

// prepareIncidentContext formats alerts into a readable context
func (c *Client) prepareIncidentContext(alerts []domain.Alert) string {
	var sb strings.Builder

	sb.WriteString("Alerts Timeline:\n")
	for i, alert := range alerts {
		sb.WriteString(fmt.Sprintf(
			"%d. [%s] %s on %s - Status: %s -> %s (Value: %.2f)\n",
			i+1,
			alert.OccurredAt.Format("15:04:05"),
			alert.Name,
			alert.Host,
			alert.OldStatus,
			alert.Status,
			alert.Value,
		))
		if alert.Description != "" {
			sb.WriteString(fmt.Sprintf("   Description: %s\n", alert.Description))
		}
	}

	sb.WriteString("\nAffected Resources:\n")
	resourceMap := make(map[string]bool)
	for _, alert := range alerts {
		if !resourceMap[alert.Host] {
			resourceMap[alert.Host] = true
			sb.WriteString(fmt.Sprintf("- Host: %s (%s)\n", alert.Host, alert.ResourceType))
		}
	}

	return sb.String()
}

// parseRecommendations parses the recommendations response
func (c *Client) parseRecommendations(response string) Recommendations {
	lines := strings.Split(response, "\n")

	rec := Recommendations{
		Immediate:  []string{},
		ShortTerm:  []string{},
		LongTerm:   []string{},
	}

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "IMMEDIATE") {
			currentSection = "immediate"
			continue
		} else if strings.HasPrefix(line, "SHORT TERM") {
			currentSection = "short_term"
			continue
		} else if strings.HasPrefix(line, "LONG TERM") {
			currentSection = "long_term"
			continue
		}

		if strings.HasPrefix(line, "-") {
			action := strings.TrimPrefix(line, "-")
			action = strings.TrimSpace(action)
			if action == "" {
				continue
			}

			switch currentSection {
			case "immediate":
				rec.Immediate = append(rec.Immediate, action)
			case "short_term":
				rec.ShortTerm = append(rec.ShortTerm, action)
			case "long_term":
				rec.LongTerm = append(rec.LongTerm, action)
			}
		}
	}

	return rec
}

// calculateTimeSpan returns the duration of the incident
func (c *Client) calculateTimeSpan(alerts []domain.Alert) time.Duration {
	if len(alerts) == 0 {
		return 0
	}

	start := alerts[0].OccurredAt
	end := alerts[len(alerts)-1].OccurredAt

	return end.Sub(start)
}

// Types

// IncidentAnalysis contains the full analysis of an incident
type IncidentAnalysis struct {
	Summary         string
	RootCause       string
	Recommendations Recommendations
	Impact          string
	GeneratedAt     time.Time
	AlertCount      int
	TimeSpan        time.Duration
}

// Recommendations contains actions organized by timeframe
type Recommendations struct {
	Immediate []string
	ShortTerm []string
	LongTerm  []string
}
