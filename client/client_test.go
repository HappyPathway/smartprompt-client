package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefinePromptWithOptions(t *testing.T) {
	domain := DomainInfrastructure
	expertLevel := ExpertiseExpert
	format := OutputFormatTutorial
	includeBest := true
	includeExamples := true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		// Verify request parameters
		if *req.Domain != domain {
			t.Errorf("Expected domain %s, got %s", domain, *req.Domain)
		}
		if *req.ExpertiseLevel != expertLevel {
			t.Errorf("Expected expertise level %s, got %s", expertLevel, *req.ExpertiseLevel)
		}
		if *req.OutputFormat != format {
			t.Errorf("Expected output format %s, got %s", format, *req.OutputFormat)
		}
		if *req.IncludeBestPractices != includeBest {
			t.Errorf("Expected includeBestPractices %v, got %v", includeBest, *req.IncludeBestPractices)
		}
		if *req.IncludeExamples != includeExamples {
			t.Errorf("Expected includeExamples %v, got %v", includeExamples, *req.IncludeExamples)
		}

		resp := PromptResponse{
			RefinedPrompt: "Enhanced: " + req.LazyPrompt,
			DetectedTopics: []string{"Topic1", "Topic2"},
			RecommendedReferences: []string{"Ref1", "Ref2"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 1*time.Second)
	response, err := client.RefinePromptWithOptions(
		"test prompt",
		&domain,
		&expertLevel,
		&format,
		&includeBest,
		&includeExamples,
	)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if response.RefinedPrompt != "Enhanced: test prompt" {
		t.Errorf("Expected 'Enhanced: test prompt', got %q", response.RefinedPrompt)
	}
	if len(response.DetectedTopics) != 2 {
		t.Errorf("Expected 2 detected topics, got %d", len(response.DetectedTopics))
	}
	if len(response.RecommendedReferences) != 2 {
		t.Errorf("Expected 2 recommended references, got %d", len(response.RecommendedReferences))
	}
}

func TestRefinePrompt(t *testing.T) {
	// Test the simple interface still works
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		// Verify optional parameters are not set
		if req.Domain != nil {
			t.Error("Expected nil domain")
		}
		if req.ExpertiseLevel != nil {
			t.Error("Expected nil expertise level")
		}
		if req.OutputFormat != nil {
			t.Error("Expected nil output format")
		}

		resp := PromptResponse{
			RefinedPrompt: "Simple: " + req.LazyPrompt,
			DetectedTopics: []string{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 1*time.Second)
	refined, err := client.RefinePrompt("test prompt")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if refined != "Simple: test prompt" {
		t.Errorf("Expected 'Simple: test prompt', got %q", refined)
	}
}

func TestRefinePromptMaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 1*time.Second)
	client.RetryConfig = &RetryConfig{
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond,
		Multiplier: 2.0,
	}

	_, err := client.RefinePrompt("test prompt")
	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}
}

func TestRefinePromptClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, 1*time.Second)
	
	_, err := client.RefinePrompt("test prompt")
	if err == nil {
		t.Fatal("Expected error for client error status, got nil")
	}
}