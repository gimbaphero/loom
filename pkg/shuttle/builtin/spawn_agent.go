// Copyright © 2026 Teradata Corporation - All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/teradata-labs/loom/pkg/shuttle"
)

// SpawnHandler is implemented by MultiAgentServer to handle agent spawning requests.
// This interface abstracts the server-side spawning logic from the tool implementation.
type SpawnHandler interface {
	// SpawnSubAgent spawns a new agent as a child of the current session
	SpawnSubAgent(ctx context.Context, req *SpawnSubAgentRequest) (*SpawnSubAgentResponse, error)
}

// SpawnSubAgentRequest contains parameters for spawning a new sub-agent.
type SpawnSubAgentRequest struct {
	ParentSessionID string            // Session ID of the parent agent
	ParentAgentID   string            // Agent ID of the parent
	AgentID         string            // Agent config to spawn (e.g., "fighter")
	WorkflowID      string            // Optional: workflow namespace (e.g., "dungeon-crawl-workflow")
	InitialMessage  string            // Optional: first message to send to spawned agent
	AutoSubscribe   []string          // Optional: topics to auto-subscribe
	Metadata        map[string]string // Optional: metadata for tracking
}

// SpawnSubAgentResponse contains the result of spawning a sub-agent.
type SpawnSubAgentResponse struct {
	SubAgentID       string   // Full agent ID (with workflow prefix if provided)
	SessionID        string   // New session ID for the sub-agent
	Status           string   // Status: "spawned", "running", "error"
	SubscribedTopics []string // Topics the agent auto-subscribed to
}

// SpawnAgentTool enables agents to dynamically spawn sub-agents.
// Use cases:
// - Interactive workflows (spawn party members on demand)
// - Context isolation (spawn fresh agent with clean context)
// - Parallel delegation (spawn multiple specialists)
// - Dynamic scaling (create agents as needed)
type SpawnAgentTool struct {
	handler       SpawnHandler
	parentSession string // Session ID of the agent using this tool
	parentAgentID string // Agent ID of the agent using this tool
}

// NewSpawnAgentTool creates a new spawn_agent tool for an agent.
func NewSpawnAgentTool(handler SpawnHandler, parentSessionID string, parentAgentID string) *SpawnAgentTool {
	return &SpawnAgentTool{
		handler:       handler,
		parentSession: parentSessionID,
		parentAgentID: parentAgentID,
	}
}

func (t *SpawnAgentTool) Name() string {
	return "spawn_agent"
}

// Description returns the tool description.
func (t *SpawnAgentTool) Description() string {
	return `Spawn a new agent instance to run in the background.

Use this tool to:
- Spawn party members for interactive workflows (e.g., spawn fighter, wizard, rogue for D&D)
- Create specialists for parallel tasks (e.g., spawn sql-analyst, security-analyst)
- Isolate context (spawn fresh agent when current context is bloated)
- Scale dynamically (create agents on demand, not all upfront)

The spawned agent runs independently in the background with its own session.
You can communicate with spawned agents via:
- Pub/sub: Use auto_subscribe to have agent join topics (e.g., "party-chat")
- Message queue: Use send_message to send direct messages

Spawned agents are automatically cleaned up when your session ends.

Examples:
- spawn_agent("fighter", workflow_id="dungeon-crawl", auto_subscribe=["party-chat"])
- spawn_agent("analyst", initial_message="Analyze this data: ...")
- spawn_agent("helper", metadata={"role": "assistant"})`
}

func (t *SpawnAgentTool) InputSchema() *shuttle.JSONSchema {
	return shuttle.NewObjectSchema(
		"Parameters for spawning a new agent",
		map[string]*shuttle.JSONSchema{
			"agent_id": shuttle.NewStringSchema("Agent config to spawn (e.g., 'fighter', 'analyst')"),
			"workflow_id": shuttle.NewStringSchema("Optional: Workflow namespace (e.g., 'dungeon-crawl-workflow')"),
			"initial_message": shuttle.NewStringSchema("Optional: First message to send to spawned agent"),
			"auto_subscribe": shuttle.NewArraySchema(
				"Optional: Topics to auto-subscribe (e.g., ['party-chat'])",
				shuttle.NewStringSchema("Topic name"),
			),
			"metadata": shuttle.NewObjectSchema(
				"Optional: Metadata key-value pairs for tracking",
				map[string]*shuttle.JSONSchema{},
				nil,
			),
		},
		[]string{"agent_id"},
	)
}

func (t *SpawnAgentTool) Execute(ctx context.Context, params map[string]interface{}) (*shuttle.Result, error) {
	start := time.Now()

	// Validate handler availability
	if t.handler == nil {
		return &shuttle.Result{
			Success: false,
			Error: &shuttle.Error{
				Code:       "SPAWN_NOT_AVAILABLE",
				Message:    "Agent spawning not configured for this server",
				Suggestion: "spawn_agent requires MultiAgentServer with agent registry",
			},
			ExecutionTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// Extract agent_id (required)
	agentID, ok := params["agent_id"].(string)
	if !ok || agentID == "" {
		return &shuttle.Result{
			Success: false,
			Error: &shuttle.Error{
				Code:       "INVALID_AGENT_ID",
				Message:    "agent_id must be a non-empty string",
				Suggestion: "Provide agent_id like 'fighter' or 'analyst'",
			},
			ExecutionTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// Extract optional workflow_id
	workflowID := ""
	if wid, ok := params["workflow_id"].(string); ok {
		workflowID = wid
	}

	// Extract optional initial_message
	initialMessage := ""
	if msg, ok := params["initial_message"].(string); ok {
		initialMessage = msg
	}

	// Extract optional auto_subscribe
	var autoSubscribe []string
	if sub, ok := params["auto_subscribe"].([]any); ok {
		for _, topic := range sub {
			if topicStr, ok := topic.(string); ok {
				autoSubscribe = append(autoSubscribe, topicStr)
			}
		}
	}

	// Extract optional metadata
	metadata := make(map[string]string)
	if md, ok := params["metadata"].(map[string]any); ok {
		for k, v := range md {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			}
		}
	}

	// Build spawn request
	req := &SpawnSubAgentRequest{
		ParentSessionID: t.parentSession,
		ParentAgentID:   t.parentAgentID,
		AgentID:         agentID,
		WorkflowID:      workflowID,
		InitialMessage:  initialMessage,
		AutoSubscribe:   autoSubscribe,
		Metadata:        metadata,
	}

	// Call server to spawn agent
	resp, err := t.handler.SpawnSubAgent(ctx, req)
	if err != nil {
		return &shuttle.Result{
			Success: false,
			Error: &shuttle.Error{
				Code:       "SPAWN_FAILED",
				Message:    fmt.Sprintf("Failed to spawn agent: %v", err),
				Retryable:  true,
				Suggestion: "Check if agent config exists and server has capacity",
			},
			ExecutionTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// Build result
	result := map[string]interface{}{
		"success":           true,
		"sub_agent_id":      resp.SubAgentID,
		"session_id":        resp.SessionID,
		"status":            resp.Status,
		"subscribed_topics": resp.SubscribedTopics,
	}

	return &shuttle.Result{
		Success: true,
		Data:    result,
		Metadata: map[string]interface{}{
			"sub_agent_id":      resp.SubAgentID,
			"session_id":        resp.SessionID,
			"subscribed_topics": resp.SubscribedTopics,
		},
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

func (t *SpawnAgentTool) Backend() string {
	return "" // Backend-agnostic
}
