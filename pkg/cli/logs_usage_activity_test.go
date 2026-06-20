package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUsageActivitySummary(t *testing.T) {
	t.Parallel()

	runDir := t.TempDir()
	summaryPath := filepath.Join(runDir, "usage", "activity", "summary.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(summaryPath), 0o755))
	require.NoError(t, os.WriteFile(summaryPath, []byte(`{
		"schema":"usage-activity-summary/v1",
		"firewall":{"total_requests":10,"allowed_requests":8,"blocked_requests":2},
		"session":{"turns":7},
		"gateway":{"total_calls":5,"failed_calls":1}
	}`), 0o644))

	summary, err := loadUsageActivitySummary(runDir)
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.NotNil(t, summary.Firewall)
	assert.Equal(t, 10, summary.Firewall.TotalRequests)
	require.NotNil(t, summary.Session)
	assert.Equal(t, 7, summary.Session.Turns)
	require.NotNil(t, summary.Gateway)
	assert.Equal(t, 5, summary.Gateway.TotalCalls)
}

func TestApplyUsageActivitySummaryToResult(t *testing.T) {
	t.Parallel()

	result := DownloadResult{}
	summary := &usageActivitySummary{
		Session: &usageActivitySession{Turns: 4},
		Firewall: &usageActivityFirewall{
			TotalRequests:   12,
			AllowedRequests: 9,
			BlockedRequests: 3,
		},
		Gateway: &usageActivityGateway{
			TotalCalls:  6,
			FailedCalls: 2,
			Servers: []usageActivityGatewayServer{
				{ServerName: "github", ToolCallCount: 5, FailedCalls: 2},
				{ServerName: "playwright", ToolCallCount: 1, FailedCalls: 0},
			},
		},
	}

	applyUsageActivitySummaryToResult(summary, &result)

	assert.Equal(t, 4, result.Run.Turns)
	require.NotNil(t, result.FirewallAnalysis)
	assert.Equal(t, 12, result.FirewallAnalysis.TotalRequests)
	assert.Equal(t, 3, result.FirewallAnalysis.BlockedRequests)
	require.NotNil(t, result.MCPToolUsage)
	require.Len(t, result.MCPToolUsage.Servers, 2)
	assert.Equal(t, "github", result.MCPToolUsage.Servers[0].ServerName)
	assert.Equal(t, 5, result.MCPToolUsage.Servers[0].ToolCallCount)
	assert.Equal(t, 2, result.MCPToolUsage.Servers[0].ErrorCount)
}

