package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type DomainType string
type ExpertiseLevel string
type OutputFormat string

const (
	DomainArchitecture    DomainType = "architecture"
	DomainDevelopment     DomainType = "development"
	DomainInfrastructure  DomainType = "infrastructure"
	DomainSecurity        DomainType = "security"
	DomainGeneral         DomainType = "general"

	ExpertiseBeginner     ExpertiseLevel = "beginner"
	ExpertiseIntermediate ExpertiseLevel = "intermediate"
	ExpertiseExpert       ExpertiseLevel = "expert"

	OutputFormatSimple    OutputFormat = "simple"
	OutputFormatDetailed  OutputFormat = "detailed"
	OutputFormatTutorial  OutputFormat = "tutorial"
	OutputFormatChecklist OutputFormat = "checklist"
)

type RetryConfig struct {
	MaxRetries int
	RetryDelay time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	RetryConfig *RetryConfig
}

type PromptRequest struct {
	LazyPrompt         string         `json:"lazy_prompt"`
	Domain            *DomainType     `json:"domain,omitempty"`
	ExpertiseLevel    *ExpertiseLevel `json:"expertise_level,omitempty"`
	OutputFormat      *OutputFormat   `json:"output_format,omitempty"`
	IncludeBestPractices *bool       `json:"include_best_practices,omitempty"`
	IncludeExamples   *bool          `json:"include_examples,omitempty"`
}

type PromptResponse struct {
	RefinedPrompt        string   `json:"refined_prompt"`
	DetectedTopics       []string `json:"detected_topics"`
	RecommendedReferences []string `json:"recommended_references,omitempty"`
}

// Default retry configuration
var DefaultRetryConfig = &RetryConfig{
	MaxRetries: 3,
	RetryDelay: 100 * time.Millisecond,
	MaxDelay:   2 * time.Second,
	Multiplier: 2.0,
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		RetryConfig: DefaultRetryConfig,
	}
}

// WithRetryConfig sets a custom retry configuration
func (c *Client) WithRetryConfig(config *RetryConfig) *Client {
	c.RetryConfig = config
	return c
}

// RefinePromptWithOptions is an enhanced version that supports all API options
func (c *Client) RefinePromptWithOptions(lazyPrompt string, domain *DomainType, expertiseLevel *ExpertiseLevel, 
	outputFormat *OutputFormat, includeBestPractices, includeExamples *bool) (*PromptResponse, error) {
	
	if lazyPrompt == "" {
		return nil, errors.New("lazy prompt cannot be empty")
	}

	reqBody := PromptRequest{
		LazyPrompt:          lazyPrompt,
		Domain:             domain,
		ExpertiseLevel:     expertiseLevel,
		OutputFormat:       outputFormat,
		IncludeBestPractices: includeBestPractices,
		IncludeExamples:    includeExamples,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	var lastErr error
	currentDelay := c.RetryConfig.RetryDelay

	for attempt := 0; attempt <= c.RetryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(currentDelay)
			currentDelay = time.Duration(float64(currentDelay) * c.RetryConfig.Multiplier)
			if currentDelay > c.RetryConfig.MaxDelay {
				currentDelay = c.RetryConfig.MaxDelay
			}
		}

		resp, err := c.HTTPClient.Post(
			c.BaseURL+"/refine-prompt",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			lastErr = fmt.Errorf("error making request (attempt %d): %w", attempt+1, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result PromptResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("error decoding response: %w", err)
			}
			resp.Body.Close()
			return &result, nil
		}

		resp.Body.Close()
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Don't retry client errors (4xx)
			return nil, fmt.Errorf("API returned client error status: %d", resp.StatusCode)
		}
		lastErr = fmt.Errorf("API returned non-200 status: %d (attempt %d)", resp.StatusCode, attempt+1)
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RefinePrompt maintains backward compatibility with the simple interface
func (c *Client) RefinePrompt(lazyPrompt string) (string, error) {
	response, err := c.RefinePromptWithOptions(lazyPrompt, nil, nil, nil, nil, nil)
	if err != nil {
		return "", err
	}
	return response.RefinedPrompt, nil
}