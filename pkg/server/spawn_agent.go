// Copyright © 2026 Teradata Corporation - All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/teradata-labs/loom/pkg/agent"
	"github.com/teradata-labs/loom/pkg/shuttle/builtin"
	"go.uber.org/zap"
)

// SpawnSubAgent spawns a new agent as a child of the current session.
// This implements the builtin.SpawnHandler interface.
func (s *MultiAgentServer) SpawnSubAgent(ctx context.Context, req *builtin.SpawnSubAgentRequest) (*builtin.SpawnSubAgentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("spawn request cannot be nil")
	}

	// Validate required fields
	if req.ParentSessionID == "" {
		return nil, fmt.Errorf("parent session ID is required")
	}
	if req.AgentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	// Check registry is available
	s.mu.RLock()
	registry := s.registry
	logger := s.logger
	messageBus := s.messageBus
	s.mu.RUnlock()

	if registry == nil {
		return nil, fmt.Errorf("agent registry not configured")
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	logger.Info("Spawning sub-agent",
		zap.String("parent_session", req.ParentSessionID),
		zap.String("parent_agent", req.ParentAgentID),
		zap.String("agent_id", req.AgentID),
		zap.String("workflow_id", req.WorkflowID))

	// Check spawn limits (prevent spawn bombs)
	existingSpawns := s.countSpawnedAgentsByParent(req.ParentSessionID)
	maxSpawnsPerParent := 10 // TODO: Make configurable
	if existingSpawns >= maxSpawnsPerParent {
		return nil, fmt.Errorf("spawn limit reached: parent has %d spawned agents (max: %d)", existingSpawns, maxSpawnsPerParent)
	}

	// Build full sub-agent ID (with workflow prefix if provided)
	subAgentID := req.AgentID
	if req.WorkflowID != "" {
		subAgentID = fmt.Sprintf("%s:%s", req.WorkflowID, req.AgentID)
	}

	// Get or load agent from registry
	ag, err := s.getOrLoadAgent(ctx, req.AgentID, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent %s: %w", req.AgentID, err)
	}

	// Create new session for sub-agent
	sessionID := GenerateSessionID()
	session := &agent.Session{
		ID:              sessionID,
		AgentID:         req.AgentID,
		ParentSessionID: req.ParentSessionID, // Link to parent
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Store session
	if err := s.sessionStore.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	logger.Info("Created sub-agent session",
		zap.String("session_id", sessionID),
		zap.String("sub_agent_id", subAgentID))

	// Auto-subscribe to topics if specified
	var subscribedTopics []string
	if messageBus != nil && len(req.AutoSubscribe) > 0 {
		for _, topic := range req.AutoSubscribe {
			subscription, err := messageBus.Subscribe(ctx, subAgentID, topic, nil, 100)
			if err != nil {
				logger.Warn("Failed to auto-subscribe to topic",
					zap.String("topic", topic),
					zap.String("sub_agent_id", subAgentID),
					zap.Error(err))
				continue
			}
			subscribedTopics = append(subscribedTopics, topic)
			logger.Info("Auto-subscribed spawned agent to topic",
				zap.String("sub_agent_id", subAgentID),
				zap.String("topic", topic),
				zap.String("subscription_id", subscription.ID))
		}
	}

	// Create context with cancel for lifecycle management
	subCtx, cancel := context.WithCancel(context.Background())

	// Track spawned agent
	spawnedAgent := &spawnedAgentContext{
		parentSessionID: req.ParentSessionID,
		parentAgentID:   req.ParentAgentID,
		subAgentID:      subAgentID,
		subSessionID:    sessionID,
		workflowID:      req.WorkflowID,
		agent:           ag,
		spawnedAt:       time.Now(),
		subscriptions:   subscribedTopics,
		metadata:        req.Metadata,
		cancelFunc:      cancel,
	}

	s.spawnedAgentsMu.Lock()
	s.spawnedAgents[sessionID] = spawnedAgent
	s.spawnedAgentsMu.Unlock()

	logger.Info("Spawned sub-agent tracked",
		zap.String("session_id", sessionID),
		zap.String("sub_agent_id", subAgentID),
		zap.Int("subscribed_topics", len(subscribedTopics)))

	// TODO: Send initial message if provided
	// For now, parent must send initial message via send_message or publish
	// The initial_message parameter is stored in metadata for future use
	if req.InitialMessage != "" {
		if spawnedAgent.metadata == nil {
			spawnedAgent.metadata = make(map[string]string)
		}
		spawnedAgent.metadata["initial_message"] = req.InitialMessage
		logger.Info("Initial message stored in metadata (parent should send via send_message/publish)",
			zap.String("session_id", sessionID),
			zap.String("message_preview", truncateString(req.InitialMessage, 50)))
	}

	// Start background monitoring for sub-agent lifecycle
	go s.monitorSpawnedAgent(subCtx, sessionID)

	// Build response
	resp := &builtin.SpawnSubAgentResponse{
		SubAgentID:       subAgentID,
		SessionID:        sessionID,
		Status:           "spawned",
		SubscribedTopics: subscribedTopics,
	}

	logger.Info("Sub-agent spawn complete",
		zap.String("sub_agent_id", subAgentID),
		zap.String("session_id", sessionID),
		zap.Int("subscribed_topics", len(subscribedTopics)))

	return resp, nil
}

// countSpawnedAgentsByParent counts how many agents a parent has spawned
func (s *MultiAgentServer) countSpawnedAgentsByParent(parentSessionID string) int {
	s.spawnedAgentsMu.RLock()
	defer s.spawnedAgentsMu.RUnlock()

	count := 0
	for _, spawned := range s.spawnedAgents {
		if spawned.parentSessionID == parentSessionID {
			count++
		}
	}
	return count
}

// monitorSpawnedAgent monitors a spawned agent's lifecycle and cleans up when done
func (s *MultiAgentServer) monitorSpawnedAgent(ctx context.Context, sessionID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger := s.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	for {
		select {
		case <-ctx.Done():
			// Context canceled (parent shutdown)
			logger.Info("Spawned agent monitor canceled",
				zap.String("session_id", sessionID))
			s.cleanupSpawnedAgent(sessionID, "parent context canceled")
			return

		case <-ticker.C:
			// Check if session is still active
			session, err := s.sessionStore.LoadSession(ctx, sessionID)
			if err != nil {
				logger.Warn("Failed to get spawned agent session",
					zap.String("session_id", sessionID),
					zap.Error(err))
				continue
			}

			// Check if session expired (no activity for 10 minutes)
			if time.Since(session.UpdatedAt) > 10*time.Minute {
				logger.Info("Spawned agent session expired",
					zap.String("session_id", sessionID),
					zap.Duration("idle_time", time.Since(session.UpdatedAt)))
				s.cleanupSpawnedAgent(sessionID, "session expired")
				return
			}
		}
	}
}

// cleanupSpawnedAgent removes a spawned agent from tracking and cleans up resources
func (s *MultiAgentServer) cleanupSpawnedAgent(sessionID string, reason string) {
	s.spawnedAgentsMu.Lock()
	spawned, exists := s.spawnedAgents[sessionID]
	if !exists {
		s.spawnedAgentsMu.Unlock()
		return
	}
	delete(s.spawnedAgents, sessionID)
	s.spawnedAgentsMu.Unlock()

	logger := s.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	logger.Info("Cleaning up spawned agent",
		zap.String("session_id", sessionID),
		zap.String("sub_agent_id", spawned.subAgentID),
		zap.String("reason", reason))

	// Cancel agent context
	if spawned.cancelFunc != nil {
		spawned.cancelFunc()
	}

	// Unsubscribe from topics
	if s.messageBus != nil {
		for _, topic := range spawned.subscriptions {
			// Get subscriptions for this agent
			subscriptions := s.messageBus.GetSubscriptionsByAgent(spawned.subAgentID)
			for _, sub := range subscriptions {
				if sub.Topic == topic {
					_ = s.messageBus.Unsubscribe(context.Background(), sub.ID)
					logger.Debug("Unsubscribed spawned agent from topic",
						zap.String("sub_agent_id", spawned.subAgentID),
						zap.String("topic", topic))
				}
			}
		}
	}

	logger.Info("Spawned agent cleanup complete",
		zap.String("session_id", sessionID),
		zap.String("sub_agent_id", spawned.subAgentID))
}

// cleanupSpawnedAgentsByParent cleans up all spawned agents for a parent session
func (s *MultiAgentServer) cleanupSpawnedAgentsByParent(parentSessionID string) {
	s.spawnedAgentsMu.Lock()
	var toCleanup []string
	for sessionID, spawned := range s.spawnedAgents {
		if spawned.parentSessionID == parentSessionID {
			toCleanup = append(toCleanup, sessionID)
		}
	}
	s.spawnedAgentsMu.Unlock()

	logger := s.logger
	if logger == nil {
		logger = zap.NewNop()
	}

	if len(toCleanup) > 0 {
		logger.Info("Cleaning up spawned agents for parent",
			zap.String("parent_session", parentSessionID),
			zap.Int("spawned_count", len(toCleanup)))

		for _, sessionID := range toCleanup {
			s.cleanupSpawnedAgent(sessionID, "parent session ended")
		}
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
