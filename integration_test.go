package toolcache

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestIntegration_EndToEnd tests the full workflow: keyer + cache + middleware + policy.
// It verifies that:
// 1. First execution with a tool and input results in a cache miss (executor called)
// 2. Second execution with same tool and input results in a cache hit (executor NOT called)
// 3. Results are identical between cache hit and miss
func TestIntegration_EndToEnd(t *testing.T) {
	// Setup: cache + keyer + policy + middleware
	cache := NewMemoryCache(DefaultPolicy())
	keyer := NewDefaultKeyer()
	policy := Policy{
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		AllowUnsafe: false,
	}
	mw := NewCacheMiddleware(cache, keyer, policy, DefaultSkipRule)

	ctx := context.Background()

	// Track executor calls
	executorCalls := 0

	// Simulate tool execution (e.g., fs:read_file)
	readFileExecutor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		executorCalls++
		inputMap := input.(map[string]any)
		path := inputMap["path"].(string)
		return []byte(fmt.Sprintf("content of %s", path)), nil
	}

	// First read: cache miss (executor should be called)
	result1, err := mw.Execute(ctx, "fs:read_file", map[string]any{"path": "/tmp/test.txt"}, []string{"read"}, readFileExecutor)
	if err != nil {
		t.Fatalf("first execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call after first execute, got %d", executorCalls)
	}
	expectedResult := []byte("content of /tmp/test.txt")
	if string(result1) != string(expectedResult) {
		t.Errorf("first result mismatch: got %q, want %q", string(result1), string(expectedResult))
	}

	// Second read (same input): cache hit (executor should NOT be called again)
	result2, err := mw.Execute(ctx, "fs:read_file", map[string]any{"path": "/tmp/test.txt"}, []string{"read"}, readFileExecutor)
	if err != nil {
		t.Fatalf("second execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call after second execute (cache hit), got %d", executorCalls)
	}
	if string(result2) != string(expectedResult) {
		t.Errorf("second result mismatch: got %q, want %q", string(result2), string(expectedResult))
	}

	// Verify results are identical
	if string(result1) != string(result2) {
		t.Errorf("results differ: first %q, second %q", string(result1), string(result2))
	}
}

// TestIntegration_MultipleTools verifies that different tools with the same input
// produce different cache entries. Each tool should be cached independently.
func TestIntegration_MultipleTools(t *testing.T) {
	// Setup
	cache := NewMemoryCache(DefaultPolicy())
	keyer := NewDefaultKeyer()
	mw := NewCacheMiddleware(cache, keyer, DefaultPolicy(), DefaultSkipRule)

	ctx := context.Background()
	tool1Calls := 0
	tool2Calls := 0

	// Tool 1 executor
	tool1 := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		tool1Calls++
		return []byte("tool1-result"), nil
	}

	// Tool 2 executor
	tool2 := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		tool2Calls++
		return []byte("tool2-result"), nil
	}

	input := map[string]any{"key": "value"}

	// Execute tool1 twice (first miss, second hit)
	result1a, err := mw.Execute(ctx, "ns:tool1", input, nil, tool1)
	if err != nil {
		t.Fatalf("tool1 first execute failed: %v", err)
	}
	if tool1Calls != 1 {
		t.Errorf("expected 1 tool1 call after first execute, got %d", tool1Calls)
	}

	result1b, err := mw.Execute(ctx, "ns:tool1", input, nil, tool1)
	if err != nil {
		t.Fatalf("tool1 second execute failed: %v", err)
	}
	if tool1Calls != 1 {
		t.Errorf("expected 1 tool1 call after second execute (cache hit), got %d", tool1Calls)
	}

	// Execute tool2 twice (first miss, second hit)
	result2a, err := mw.Execute(ctx, "ns:tool2", input, nil, tool2)
	if err != nil {
		t.Fatalf("tool2 first execute failed: %v", err)
	}
	if tool2Calls != 1 {
		t.Errorf("expected 1 tool2 call after first execute, got %d", tool2Calls)
	}

	result2b, err := mw.Execute(ctx, "ns:tool2", input, nil, tool2)
	if err != nil {
		t.Fatalf("tool2 second execute failed: %v", err)
	}
	if tool2Calls != 1 {
		t.Errorf("expected 1 tool2 call after second execute (cache hit), got %d", tool2Calls)
	}

	// Verify each tool was called exactly once (cache hits on second calls)
	if tool1Calls != 1 {
		t.Errorf("tool1 should be called exactly once, got %d", tool1Calls)
	}
	if tool2Calls != 1 {
		t.Errorf("tool2 should be called exactly once, got %d", tool2Calls)
	}

	// Verify results are correct and different
	if string(result1a) != "tool1-result" {
		t.Errorf("tool1 result mismatch: got %q, want %q", string(result1a), "tool1-result")
	}
	if string(result1b) != "tool1-result" {
		t.Errorf("tool1 cached result mismatch: got %q, want %q", string(result1b), "tool1-result")
	}
	if string(result2a) != "tool2-result" {
		t.Errorf("tool2 result mismatch: got %q, want %q", string(result2a), "tool2-result")
	}
	if string(result2b) != "tool2-result" {
		t.Errorf("tool2 cached result mismatch: got %q, want %q", string(result2b), "tool2-result")
	}

	// Verify results are different between tools
	if string(result1a) == string(result2a) {
		t.Errorf("tool1 and tool2 results should differ, both got %q", string(result1a))
	}
}

// TestIntegration_UnsafeTagsSkipCache verifies that tools with unsafe tags
// are not cached when AllowUnsafe is false.
func TestIntegration_UnsafeTagsSkipCache(t *testing.T) {
	cache := NewMemoryCache(DefaultPolicy())
	keyer := NewDefaultKeyer()
	policy := Policy{
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		AllowUnsafe: false,
	}
	mw := NewCacheMiddleware(cache, keyer, policy, DefaultSkipRule)

	ctx := context.Background()
	executorCalls := 0

	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		executorCalls++
		return []byte("result"), nil
	}

	input := map[string]any{"key": "value"}

	// Execute with "write" tag (unsafe) - should skip cache
	_, err := mw.Execute(ctx, "db:update", input, []string{"write"}, executor)
	if err != nil {
		t.Fatalf("first execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call, got %d", executorCalls)
	}

	// Second execute with same input and unsafe tag - should NOT use cache
	_, err = mw.Execute(ctx, "db:update", input, []string{"write"}, executor)
	if err != nil {
		t.Fatalf("second execute failed: %v", err)
	}
	if executorCalls != 2 {
		t.Errorf("expected 2 executor calls (cache skipped for unsafe), got %d", executorCalls)
	}
}

// TestIntegration_AllowUnsafePolicy verifies that when AllowUnsafe is true,
// unsafe tools are cached.
func TestIntegration_AllowUnsafePolicy(t *testing.T) {
	cache := NewMemoryCache(DefaultPolicy())
	keyer := NewDefaultKeyer()
	policy := Policy{
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		AllowUnsafe: true, // Allow caching of unsafe operations
	}
	mw := NewCacheMiddleware(cache, keyer, policy, DefaultSkipRule)

	ctx := context.Background()
	executorCalls := 0

	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		executorCalls++
		return []byte("result"), nil
	}

	input := map[string]any{"key": "value"}

	// Execute with "delete" tag (unsafe) - should cache because AllowUnsafe=true
	_, err := mw.Execute(ctx, "db:delete", input, []string{"delete"}, executor)
	if err != nil {
		t.Fatalf("first execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call, got %d", executorCalls)
	}

	// Second execute with same input and unsafe tag - should use cache
	_, err = mw.Execute(ctx, "db:delete", input, []string{"delete"}, executor)
	if err != nil {
		t.Fatalf("second execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call (cache hit), got %d", executorCalls)
	}
}

// TestIntegration_DifferentInputsDifferentCacheEntries verifies that
// different inputs produce different cache entries for the same tool.
func TestIntegration_DifferentInputsDifferentCacheEntries(t *testing.T) {
	cache := NewMemoryCache(DefaultPolicy())
	keyer := NewDefaultKeyer()
	mw := NewCacheMiddleware(cache, keyer, DefaultPolicy(), DefaultSkipRule)

	ctx := context.Background()
	executorCalls := 0

	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		executorCalls++
		inputMap := input.(map[string]any)
		value := inputMap["value"].(string)
		return []byte(fmt.Sprintf("result-%s", value)), nil
	}

	// Execute with input1
	result1, err := mw.Execute(ctx, "tool:process", map[string]any{"value": "input1"}, nil, executor)
	if err != nil {
		t.Fatalf("first execute failed: %v", err)
	}
	if executorCalls != 1 {
		t.Errorf("expected 1 executor call, got %d", executorCalls)
	}

	// Execute with input2 (different input)
	result2, err := mw.Execute(ctx, "tool:process", map[string]any{"value": "input2"}, nil, executor)
	if err != nil {
		t.Fatalf("second execute failed: %v", err)
	}
	if executorCalls != 2 {
		t.Errorf("expected 2 executor calls (different inputs), got %d", executorCalls)
	}

	// Execute with input1 again (should hit cache)
	result1Again, err := mw.Execute(ctx, "tool:process", map[string]any{"value": "input1"}, nil, executor)
	if err != nil {
		t.Fatalf("third execute failed: %v", err)
	}
	if executorCalls != 2 {
		t.Errorf("expected 2 executor calls (cache hit for input1), got %d", executorCalls)
	}

	// Verify results
	if string(result1) != "result-input1" {
		t.Errorf("result1 mismatch: got %q, want %q", string(result1), "result-input1")
	}
	if string(result2) != "result-input2" {
		t.Errorf("result2 mismatch: got %q, want %q", string(result2), "result-input2")
	}
	if string(result1Again) != "result-input1" {
		t.Errorf("result1Again mismatch: got %q, want %q", string(result1Again), "result-input1")
	}
}
