//go:build !integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAutoUpdateWorkflow_Enabled(t *testing.T) {
	dir := t.TempDir()

	err := GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir,
		Enabled:     true,
		RepoSlug:    "owner/repo",
	})
	require.NoError(t, err, "GenerateAutoUpdateWorkflow should succeed when enabled")

	outputPath := filepath.Join(dir, AutoUpdateWorkflowFileName)
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err, "agentic-update.yml should be written")

	content := string(data)
	assert.Contains(t, content, "name: Agentic Update", "should include workflow name")
	assert.Contains(t, content, "cron:", "should include cron schedule")
	assert.Contains(t, content, "Weekly (auto-update)", "should include schedule comment")
	assert.Contains(t, content, "workflow_dispatch:", "should include workflow_dispatch trigger")
	assert.Contains(t, content, "uses: ./.github/workflows/agentics-maintenance.yml", "should call maintenance workflow")
	assert.Contains(t, content, "operation: update", "should pass update operation")
	assert.Contains(t, content, "actions: write", "should grant actions: write")
	assert.Contains(t, content, "contents: write", "should grant contents: write")
	assert.Contains(t, content, "pull-requests: write", "should grant pull-requests: write")
}

func TestGenerateAutoUpdateWorkflow_Disabled(t *testing.T) {
	dir := t.TempDir()

	err := GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir,
		Enabled:     false,
	})
	require.NoError(t, err, "GenerateAutoUpdateWorkflow should succeed when disabled")

	outputPath := filepath.Join(dir, AutoUpdateWorkflowFileName)
	_, err = os.Stat(outputPath)
	assert.True(t, os.IsNotExist(err), "agentic-update.yml should not be created when disabled")
}

func TestGenerateAutoUpdateWorkflow_DisabledDeletesExistingFile(t *testing.T) {
	dir := t.TempDir()

	// Create an existing file to simulate a previously-generated workflow.
	outputPath := filepath.Join(dir, AutoUpdateWorkflowFileName)
	require.NoError(t, os.WriteFile(outputPath, []byte("old content"), 0o644))

	err := GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir,
		Enabled:     false,
	})
	require.NoError(t, err, "GenerateAutoUpdateWorkflow should succeed when disabled")

	_, err = os.Stat(outputPath)
	assert.True(t, os.IsNotExist(err), "existing agentic-update.yml should be deleted when disabled")
}

func TestGenerateAutoUpdateWorkflow_CronIsDeterministic(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	opts := GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir1,
		Enabled:     true,
		RepoSlug:    "myorg/myrepo",
	}
	require.NoError(t, GenerateAutoUpdateWorkflow(opts))

	opts.WorkflowDir = dir2
	require.NoError(t, GenerateAutoUpdateWorkflow(opts))

	data1, err := os.ReadFile(filepath.Join(dir1, AutoUpdateWorkflowFileName))
	require.NoError(t, err)
	data2, err := os.ReadFile(filepath.Join(dir2, AutoUpdateWorkflowFileName))
	require.NoError(t, err)

	assert.Equal(t, string(data1), string(data2), "same repo slug should produce identical output")
}

func TestGenerateAutoUpdateWorkflow_DifferentReposDifferentCron(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir1,
		Enabled:     true,
		RepoSlug:    "org1/repo-alpha",
	}))
	require.NoError(t, GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir2,
		Enabled:     true,
		RepoSlug:    "org2/repo-beta",
	}))

	data1, err := os.ReadFile(filepath.Join(dir1, AutoUpdateWorkflowFileName))
	require.NoError(t, err)
	data2, err := os.ReadFile(filepath.Join(dir2, AutoUpdateWorkflowFileName))
	require.NoError(t, err)

	// Extract cron lines and compare — different repos should (almost certainly) scatter differently.
	cron1 := extractCronLine(string(data1))
	cron2 := extractCronLine(string(data2))
	assert.NotEmpty(t, cron1, "cron should be non-empty for org1/repo-alpha")
	assert.NotEmpty(t, cron2, "cron should be non-empty for org2/repo-beta")
	// Schedules are scattered by hash — different repos should typically differ.
	// This is a best-effort check; hash collisions are possible but unlikely for these slugs.
	assert.NotEqual(t, cron1, cron2, "different repo slugs should produce different cron schedules")
}

func TestGenerateAutoUpdateWorkflow_NoRepoSlug(t *testing.T) {
	dir := t.TempDir()

	err := GenerateAutoUpdateWorkflow(GenerateAutoUpdateWorkflowOptions{
		WorkflowDir: dir,
		Enabled:     true,
		RepoSlug:    "",
	})
	require.NoError(t, err, "GenerateAutoUpdateWorkflow should succeed with empty repo slug")

	content, err := os.ReadFile(filepath.Join(dir, AutoUpdateWorkflowFileName))
	require.NoError(t, err)
	assert.Contains(t, string(content), "cron:", "should still generate a cron schedule without repo slug")
}

func TestBuildAutoUpdateSeed(t *testing.T) {
	assert.Equal(t, "owner/repo/agentic-update", buildAutoUpdateSeed("owner/repo"))
	assert.Equal(t, "agentic-update", buildAutoUpdateSeed(""))
}

// extractCronLine returns the cron expression from the first `- cron:` line in the YAML.
func extractCronLine(content string) string {
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- cron:") {
			return trimmed
		}
	}
	return ""
}
