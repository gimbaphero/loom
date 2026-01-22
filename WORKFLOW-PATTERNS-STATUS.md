# Loom Workflow Patterns - Status and Testing

## ✅ Orchestration Pattern Format Fixed (Jan 22, 2026)

**Issue:** Examples in `examples/reference/workflows/` used `spec.pattern` (old format), but the implementation expects `spec.type` (new format).

**Fix Applied:** All 6 orchestration pattern examples updated to use `spec.type` format (commit e832632).

**Reference:** See `examples/workflow-all-fields-example.yaml` lines 200-370 for correct orchestration pattern format.

---

## Overview

Loom supports two categories of workflow patterns:
1. **Communication Patterns** - Real-time multi-agent coordination (uses `spec.entrypoint`)
2. **Orchestration Patterns** - Static execution flows (uses `spec.type`)

## 1. Communication Patterns (✅ Fully Working)

These workflows enable real-time agent communication using Loom's tri-modal communication system.

### Pattern A: Pub-Sub (Peer-to-Peer)

**Status:** ✅ Fully implemented and tested

**Examples:**
- `dungeon-crawl-workflow` - D&D party adventure
- `brainstorm-session` - Creative brainstorming with facilitator, creative, and analyst

**How It Works:**
- All agents subscribe to shared topic (e.g., "party-chat", "brainstorm-chat")
- Agents publish messages using `publish(topic="...", message="...")`
- Messages auto-injected via event-driven broadcast bus
- No manual polling required - responses appear automatically

**YAML Structure:**
```yaml
spec:
  entrypoint: coordinator-agent-name
  agents:
    - name: coordinator
      agent: coordinator-agent-id
    - name: participant1
      agent: participant1-agent-id
  communication:
    pattern: "peer-to-peer-pub-sub"
    topic: "shared-topic-name"
```

**Testing:**
```bash
./bin/loom chat --thread dungeon-crawl-workflow "DM, start an adventure"
./bin/loom chat --thread brainstorm-session "Brainstorm AI app names"
```

**Test Results:** ✅ All passing (tested Jan 22, 2026)

---

### Pattern B: Hub-and-Spoke (Message Queue)

**Status:** ✅ Fully implemented and tested

**Examples:**
- `dnd-campaign-workflow` - Campaign creation with coordinator + specialists
- `vacation-planning-workflow` - Vacation planning with coordinator + analysts

**How It Works:**
- Coordinator agent orchestrates specialist sub-agents
- Uses `send_message(to_agent="...", message="...")` for direct communication
- Responses auto-injected into coordinator conversation
- Event-driven - no manual polling

**YAML Structure:**
```yaml
spec:
  entrypoint: coordinator
  agents:
    - name: coordinator
      agent: coordinator-agent-id
      metadata:
        role: coordinator
        workflow: workflow-name
    - name: specialist1
      agent: specialist1-agent-id
    - name: specialist2
      agent: specialist2-agent-id
```

**Testing:**
```bash
./bin/loom chat --thread dnd-campaign-workflow "Create a pirate campaign"
./bin/loom chat --thread vacation-planning-workflow "Plan a week in Japan"
```

**Test Results:** ✅ All passing (tested Jan 22, 2026)

---

## 2. Orchestration Patterns (YAML Format Issue Found)

These workflows use static execution patterns for structured agent coordination.

**⚠️ IMPORTANT:** The examples in `examples/reference/workflows/` use `spec.pattern` (old format), but the implementation expects `spec.type` (new format).

**Reference:** See `examples/workflow-all-fields-example.yaml` for correct format.

### Pattern 1: Pipeline

**Status:** ✅ Fixed (commit e832632)

**Example:** `examples/reference/workflows/feature-pipeline.yaml`

**Description:** Sequential execution where each stage's output becomes next stage's input

**YAML Structure:**
```yaml
spec:
  type: pipeline  # ✅ CORRECT FORMAT
  initial_prompt: "Design auth system"
  stages:
    - agent_id: spec-writer
      prompt_template: "Write spec: {{previous}}"
    - agent_id: implementer
      prompt_template: "Implement: {{previous}}"
  pass_full_history: true
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

---

### Pattern 2: Fork-Join

**Status:** ✅ Fixed (commit e832632)

**Example:** `examples/reference/workflows/code-review.yaml`

**Description:** Agents execute in parallel on same prompt, results merged

**YAML Structure:**
```yaml
spec:
  type: fork_join  # ✅ CORRECT FORMAT
  prompt: "Review this code..."
  agent_ids:
    - quality-reviewer
    - security-reviewer
  merge_strategy: concatenate
  timeout_seconds: 300
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

---

### Pattern 3: Parallel

**Status:** ✅ Fixed (commit e832632)

**Examples:**
- `doc-generation.yaml`
- `security-analysis.yaml`

**Description:** Independent tasks execute in parallel with agent-specific prompts

**YAML Structure:**
```yaml
spec:
  type: parallel  # ✅ CORRECT FORMAT
  merge_strategy: concatenate
  timeout_seconds: 600
  agents:
    - id: task1
      name: Task 1 Agent
      system_prompt: |
        Do task 1...
      prompt_template: "Task 1: {{user_query}}"
    - id: task2
      name: Task 2 Agent
      system_prompt: |
        Do task 2...
      prompt_template: "Task 2: {{user_query}}"
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

---

### Pattern 4: Debate

**Status:** ✅ Fixed (commit e832632)

**Example:** `architecture-debate.yaml`

**Description:** Multiple agents debate and reach consensus through structured rounds

**YAML Structure:**
```yaml
spec:
  type: debate  # ✅ CORRECT FORMAT
  rounds: 3
  agents:
    - id: advocate
      name: Architect Advocate
      role: debater
      system_prompt: |
        Advocate for best practices...
    - id: pragmatist
      name: Pragmatist
      role: debater
      system_prompt: |
        Advocate for pragmatism...
    - id: moderator
      name: Senior Architect
      role: moderator
      system_prompt: |
        Moderate and synthesize...
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

---

### Pattern 5: Conditional

**Status:** ✅ Fixed (commit e832632)

**Example:** `complexity-routing.yaml`

**Description:** Routes execution based on classifier agent's decision

**YAML Structure:**
```yaml
spec:
  type: conditional  # ✅ CORRECT FORMAT
  agents:
    - id: classifier
      name: Complexity Classifier
      role: classifier
      system_prompt: |
        Classify as: simple, medium, complex
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

**Limitation:** Nested workflows not yet supported

---

### Pattern 6: Swarm

**Status:** ✅ Fixed (commit e832632)

**Example:** `technology-swarm.yaml`

**Description:** Collective decision-making through voting

**YAML Structure:**
```yaml
spec:
  type: swarm  # ✅ CORRECT FORMAT
  strategy: majority  # Options: majority, supermajority, unanimous
  confidence_threshold: 0.7
  share_votes: false
  agents:
    - id: expert1
      name: Database Expert
      system_prompt: |
        Evaluate database options...
    - id: expert2
      name: Performance Engineer
      system_prompt: |
        Analyze performance...
    - id: judge
      name: Senior Architect
      role: judge
      system_prompt: |
        Break ties and synthesize...
```

**Format Fixed:** ✅ Now uses `spec.type` instead of `spec.pattern`

---

## Summary

| Pattern | Type | Status | Example | Format Fixed |
|---------|------|--------|---------|--------------|
| Pub-Sub | Communication | ✅ Working | dungeon-crawl, brainstorm-session | N/A |
| Hub-and-Spoke | Communication | ✅ Working | dnd-campaign, vacation-planner | N/A |
| Pipeline | Orchestration | ✅ Format Fixed | feature-pipeline.yaml | ✅ Yes |
| Fork-Join | Orchestration | ✅ Format Fixed | code-review.yaml | ✅ Yes |
| Parallel | Orchestration | ✅ Format Fixed | doc-generation.yaml | ✅ Yes |
| Debate | Orchestration | ✅ Format Fixed | architecture-debate.yaml | ✅ Yes |
| Conditional | Orchestration | ✅ Format Fixed | complexity-routing.yaml | ✅ Yes |
| Swarm | Orchestration | ✅ Format Fixed | technology-swarm.yaml | ✅ Yes |

## Testing Status

### ✅ Communication Patterns - Tested and Working (2/8)
1. Pub-Sub pattern - Tested successfully Jan 22, 2026
2. Hub-and-Spoke pattern - Tested successfully Jan 22, 2026

### ✅ Orchestration Patterns - Format Fixed (6/8)
1. Pipeline pattern - Fixed Jan 22, 2026 (commit e832632)
2. Fork-Join pattern - Fixed Jan 22, 2026 (commit e832632)
3. Parallel pattern - Fixed Jan 22, 2026 (commit e832632)
4. Debate pattern - Fixed Jan 22, 2026 (commit e832632)
5. Conditional pattern - Fixed Jan 22, 2026 (commit e832632)
6. Swarm pattern - Fixed Jan 22, 2026 (commit e832632)

**Status:** All orchestration pattern examples now use correct `spec.type` format. Ready for CLI execution testing with `looms workflow run`.

## Next Steps

1. ✅ **Communication patterns** - Fully working and tested (Jan 22, 2026)

2. ✅ **Orchestration patterns format fix** - Complete (Jan 22, 2026, commit e832632):
   - **Fixed:** All 6 orchestration pattern examples now use `spec.type` format
   - **Files updated:**
     - `examples/reference/workflows/feature-pipeline.yaml` → `type: pipeline`
     - `examples/reference/workflows/code-review.yaml` → `type: fork_join`
     - `examples/reference/workflows/doc-generation.yaml` → `type: parallel`
     - `examples/reference/workflows/security-analysis.yaml` → `type: parallel`
     - `examples/reference/workflows/architecture-debate.yaml` → `type: debate`
     - `examples/reference/workflows/complexity-routing.yaml` → `type: conditional`
     - `examples/reference/workflows/technology-swarm.yaml` → `type: swarm`

3. **TODO: Test orchestration patterns with CLI**
   ```bash
   # Test each pattern with workflow run command:
   looms workflow run examples/reference/workflows/feature-pipeline.yaml
   looms workflow run examples/reference/workflows/code-review.yaml --prompt "$(cat test-file.go)"
   looms workflow run examples/reference/workflows/doc-generation.yaml
   # ... etc for all 6 patterns
   ```

4. **TODO: Document orchestration pattern execution results**
   - Create test cases for each pattern
   - Document any CLI usage patterns
   - Update status file with test results

## Recent Changes

**Jan 22, 2026:**
- Simplified all workflow agent prompts (50-70% reduction)
- Implemented dynamic communication injection
- All communication patterns tested and working (commit d5f3ec7)
- **Fixed all 6 orchestration pattern examples** to use `spec.type` instead of `spec.pattern` (commit e832632)
- Updated WORKFLOW-PATTERNS-STATUS.md documentation with fix status
- All examples now ready for CLI testing with `looms workflow run`

---

**Documentation:** See `examples/reference/workflows/README.md` for full workflow pattern documentation.
