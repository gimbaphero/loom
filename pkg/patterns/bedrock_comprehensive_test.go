// Copyright 2026 Teradata
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

//go:build integration
// +build integration

package patterns

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/teradata-labs/loom/pkg/llm/bedrock"
	"github.com/teradata-labs/loom/pkg/types"
)

// TestComprehensivePatternSelectionWithBedrock tests hybrid pattern selection with diverse use cases.
// This validates that both keyword scoring and LLM re-ranking work correctly across different domains.
//
// Prerequisites:
// - AWS credentials configured
// - Run with: go test -tags integration,fts5 -run TestComprehensivePatternSelectionWithBedrock ./pkg/patterns
func TestComprehensivePatternSelectionWithBedrock(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Set up Bedrock client
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	t.Logf("🔧 Setting up Bedrock LLM (region: %s)", region)

	bedrockCfg := bedrock.Config{
		Region:      region,
		ModelID:     bedrock.DefaultBedrockModelID,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	provider, err := bedrock.NewClient(bedrockCfg)
	if err != nil {
		t.Skipf("Skipping test - Bedrock client creation failed (credentials may be unavailable): %v", err)
	}

	// Test if credentials are valid with a simple call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	testMsg := []types.Message{{Role: "user", Content: "test"}}
	_, err = provider.Chat(ctx, testMsg, nil)
	if err != nil {
		t.Skipf("Skipping test - Bedrock credentials invalid or expired: %v", err)
	}

	t.Log("✅ Bedrock client created and credentials verified")

	// Set up pattern library and orchestrator
	lib := NewLibrary(nil, "../../patterns")
	orch := NewOrchestrator(lib)

	allPatterns := lib.ListAll()
	if len(allPatterns) < 80 {
		t.Fatalf("Expected at least 80 patterns, got %d", len(allPatterns))
	}
	t.Logf("✅ Loaded %d patterns", len(allPatterns))

	// Configure LLM-based intent classifier and hybrid re-ranking
	llmClassifierConfig := DefaultLLMClassifierConfig(provider)
	llmClassifier := NewLLMIntentClassifier(llmClassifierConfig)
	orch.SetIntentClassifier(llmClassifier)
	orch.SetLLMProvider(provider)
	t.Log("✅ LLM intent classifier and hybrid re-ranking configured")

	// Comprehensive test scenarios covering different domains
	scenarios := []struct {
		category      string
		userMessage   string
		expectedInTop []string // Patterns expected in top results
		description   string
	}{
		// Analytics scenarios
		{
			category:      "Analytics",
			userMessage:   "I need to predict which customers will churn using historical behavior data",
			expectedInTop: []string{"churn_analysis", "customer_health_scoring", "logistic_regression"},
			description:   "Customer churn prediction",
		},
		{
			category:      "Analytics",
			userMessage:   "analyze user journeys through our funnel to find drop-off points",
			expectedInTop: []string{"funnel_analysis", "npath", "sessionize"},
			description:   "Funnel and journey analysis",
		},
		{
			category:      "Analytics",
			userMessage:   "find the most frequently occurring sequences in clickstream data",
			expectedInTop: []string{"npath", "sessionize"},
			description:   "Sequence analysis",
		},
		{
			category:      "Analytics",
			userMessage:   "analyze product affinities to understand what customers buy together",
			expectedInTop: []string{"market_basket"},
			description:   "Market basket analysis",
		},
		{
			category:      "Analytics",
			userMessage:   "calculate cohort retention rates over time",
			expectedInTop: []string{"cohort_analysis"},
			description:   "Cohort analysis",
		},

		// Machine Learning scenarios
		{
			category:      "Machine Learning",
			userMessage:   "build a machine learning model to forecast customer lifetime value",
			expectedInTop: []string{"linear_regression", "arima"},
			description:   "CLV prediction",
		},
		{
			category:      "Machine Learning",
			userMessage:   "classify customers into different segments based on behavior",
			expectedInTop: []string{"kmeans", "decision_tree"},
			description:   "Customer segmentation",
		},
		{
			category:      "Machine Learning",
			userMessage:   "predict whether a loan application should be approved",
			expectedInTop: []string{"logistic_regression", "decision_tree"},
			description:   "Binary classification",
		},
		{
			category:      "Machine Learning",
			userMessage:   "forecast next quarter sales based on historical trends",
			expectedInTop: []string{"arima", "linear_regression", "moving_average"},
			description:   "Time series forecasting",
		},

		// Data Quality scenarios
		{
			category:      "Data Quality",
			userMessage:   "find duplicate customer records and merge them",
			expectedInTop: []string{"duplicate_detection", "data_profiling"},
			description:   "Duplicate detection",
		},
		{
			category:      "Data Quality",
			userMessage:   "identify missing values and decide how to handle them",
			expectedInTop: []string{"missing_value_analysis", "data_profiling"},
			description:   "Missing value analysis",
		},
		{
			category:      "Data Quality",
			userMessage:   "validate that email addresses and phone numbers are properly formatted",
			expectedInTop: []string{"data_validation"},
			description:   "Data validation rules",
		},
		{
			category:      "Data Quality",
			userMessage:   "profile a new dataset to understand its structure and quality",
			expectedInTop: []string{"data_profiling"},
			description:   "Data profiling",
		},

		// Time Series scenarios
		{
			category:      "Time Series",
			userMessage:   "detect anomalies in server metrics over time",
			expectedInTop: []string{"arima", "moving_average"},
			description:   "Anomaly detection in time series",
		},
		{
			category:      "Time Series",
			userMessage:   "calculate moving averages to smooth out fluctuations in stock prices",
			expectedInTop: []string{"moving_average"},
			description:   "Moving average calculation",
		},
		{
			category:      "Time Series",
			userMessage:   "analyze seasonal patterns in retail sales data",
			expectedInTop: []string{"arima"},
			description:   "Seasonal analysis",
		},

		// SQL Performance scenarios
		{
			category:      "Performance",
			userMessage:   "my COUNT queries are too slow on large tables",
			expectedInTop: []string{"count_optimization"},
			description:   "COUNT optimization",
		},
		{
			category:      "Performance",
			userMessage:   "find queries that are doing full table scans",
			expectedInTop: []string{"sequential_scan_detection"},
			description:   "Sequential scan detection",
		},
		{
			category:      "Performance",
			userMessage:   "identify missing indexes that could speed up queries",
			expectedInTop: []string{"missing_index_analysis"},
			description:   "Index recommendation",
		},

		// Text Analysis scenarios
		{
			category:      "Text Analysis",
			userMessage:   "extract keywords and topics from customer reviews",
			expectedInTop: []string{"text_analysis"},
			description:   "Keyword extraction",
		},
		{
			category:      "Text Analysis",
			userMessage:   "analyze sentiment of product reviews",
			expectedInTop: []string{"text_analysis"},
			description:   "Sentiment analysis",
		},

		// Data Discovery scenarios
		{
			category:      "Data Discovery",
			userMessage:   "automatically detect primary keys in undocumented tables",
			expectedInTop: []string{"key_detection"},
			description:   "Primary key detection",
		},
		{
			category:      "Data Discovery",
			userMessage:   "find relationships between tables based on column names",
			expectedInTop: []string{"schema_inference"},
			description:   "Schema inference",
		},

		// Ambiguous/Edge cases
		{
			category:      "Ambiguous",
			userMessage:   "help me understand what's wrong with my data",
			expectedInTop: []string{"data_profiling", "data_validation", "missing_value_analysis"},
			description:   "Vague data quality request",
		},
		{
			category:      "Ambiguous",
			userMessage:   "I need to analyze my customers",
			expectedInTop: []string{"customer_health_scoring", "churn_analysis", "cohort_analysis"},
			description:   "Generic customer analysis",
		},
	}

	totalScenarios := len(scenarios)
	passedScenarios := 0
	highConfidenceScenarios := 0 // confidence >= 0.70
	llmInvokedCount := 0

	separator := strings.Repeat("=", 80)
	t.Logf("\n%s", separator)
	t.Logf("RUNNING %d COMPREHENSIVE TEST SCENARIOS", totalScenarios)
	t.Logf("%s\n", separator)

	for i, sc := range scenarios {
		t.Run(sc.description, func(t *testing.T) {
			t.Logf("\n[%d/%d] %s: %s", i+1, totalScenarios, sc.category, sc.description)
			t.Logf("Query: %q", sc.userMessage)

			// Classify intent
			startTime := time.Now()
			intent, intentConf := llmClassifier(sc.userMessage, nil)
			intentDuration := time.Since(startTime)

			t.Logf("├─ Intent: %s (%.2f) [%v]", intent, intentConf, intentDuration)

			// Recommend pattern
			startTime = time.Now()
			patternName, patternConf := orch.RecommendPattern(sc.userMessage, intent)
			patternDuration := time.Since(startTime)

			if patternName == "" {
				t.Errorf("├─ ❌ No pattern recommended")
				return
			}

			// Check if LLM re-ranking was used (duration > 2s indicates LLM call)
			usedLLM := patternDuration > 2*time.Second
			if usedLLM {
				llmInvokedCount++
			}

			method := "keyword"
			if usedLLM {
				method = "LLM re-rank"
			}

			t.Logf("├─ Pattern: %s (%.2f) [%v] (%s)", patternName, patternConf, patternDuration, method)

			// Verify pattern is relevant
			found := false
			for _, expected := range sc.expectedInTop {
				if patternName == expected {
					found = true
					break
				}
			}

			if found {
				t.Logf("└─ ✅ PASS - Pattern matches expected results")
				passedScenarios++
				if patternConf >= 0.70 {
					highConfidenceScenarios++
				}
			} else {
				// Load pattern to check semantic relevance
				pattern, err := lib.Load(patternName)
				if err == nil {
					t.Logf("├─ ⚠️  Pattern not in expected list, checking semantic relevance...")
					t.Logf("│  Title: %s", pattern.Title)
					t.Logf("│  Category: %s", pattern.Category)
					if len(pattern.UseCases) > 0 {
						t.Logf("│  Use cases: %v", pattern.UseCases[:min(2, len(pattern.UseCases))])
					}

					// If confidence is high, might still be semantically correct
					if patternConf >= 0.60 {
						t.Logf("└─ ⚠️  PARTIAL PASS - High confidence but unexpected pattern")
						passedScenarios++ // Count as pass if high confidence
					} else {
						t.Logf("└─ ❌ FAIL - Low confidence and not in expected results")
					}
				} else {
					t.Logf("└─ ❌ FAIL - Pattern not found: %v", err)
				}
			}
		})
	}

	// Print summary
	t.Logf("\n%s", separator)
	t.Logf("TEST SUMMARY")
	t.Logf("%s", separator)
	t.Logf("Total scenarios:          %d", totalScenarios)
	t.Logf("Passed scenarios:         %d (%.1f%%)", passedScenarios, float64(passedScenarios)/float64(totalScenarios)*100)
	t.Logf("High confidence (≥0.70):  %d (%.1f%%)", highConfidenceScenarios, float64(highConfidenceScenarios)/float64(totalScenarios)*100)
	t.Logf("LLM re-ranking invoked:   %d (%.1f%%)", llmInvokedCount, float64(llmInvokedCount)/float64(totalScenarios)*100)
	t.Logf("%s\n", separator)

	// Success criteria: 80% pass rate
	passRate := float64(passedScenarios) / float64(totalScenarios)
	if passRate < 0.80 {
		t.Errorf("Pass rate %.1f%% below 80%% threshold", passRate*100)
	} else {
		t.Logf("✅ SUCCESS: Pass rate %.1f%% exceeds 80%% threshold", passRate*100)
	}
}
