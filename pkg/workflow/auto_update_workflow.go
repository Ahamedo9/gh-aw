package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/fileutil"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/parser"
)

var autoUpdateWorkflowLog = logger.New("workflow:auto_update_workflow")

// AutoUpdateWorkflowFileName is the filename for the generated auto-update workflow.
const AutoUpdateWorkflowFileName = "agentic-update.yml"

// autoUpdateWorkflowIdentifier is the stable identifier used to scatter the
// FUZZY:WEEKLY cron schedule. It is combined with the repo slug to ensure
// that different repositories scatter to different time slots.
const autoUpdateWorkflowIdentifier = "agentic-update"

// GenerateAutoUpdateWorkflowOptions configures an auto-update workflow generation run.
type GenerateAutoUpdateWorkflowOptions struct {
	// WorkflowDir is the directory where the workflow file will be written.
	WorkflowDir string
	// Enabled indicates whether auto-updates are enabled in the repo config.
	Enabled bool
	// RepoSlug is the owner/repo slug used to deterministically scatter the
	// weekly cron schedule across different repositories. Pass an empty string
	// when the slug is not available; scattering will still succeed using only
	// the workflow identifier as seed.
	RepoSlug string
}

// GenerateAutoUpdateWorkflow generates or removes the agentic-update.yml workflow
// based on whether auto-updates are enabled in the repository's aw.json.
//
// When enabled, it generates a workflow that runs on a fuzzy weekly schedule
// and dispatches the 'update' operation to agentics-maintenance.yml via workflow_call.
//
// When disabled (or when maintenance is disabled), any existing agentic-update.yml
// is deleted.
func GenerateAutoUpdateWorkflow(opts GenerateAutoUpdateWorkflowOptions) error {
	outputFile := filepath.Join(opts.WorkflowDir, AutoUpdateWorkflowFileName)

	if !opts.Enabled {
		autoUpdateWorkflowLog.Print("Auto-updates not enabled, removing agentic-update.yml if present")
		if _, err := os.Stat(outputFile); err == nil {
			autoUpdateWorkflowLog.Printf("Deleting existing auto-update workflow: %s", outputFile)
			if err := os.Remove(outputFile); err != nil {
				return fmt.Errorf("failed to delete auto-update workflow: %w", err)
			}
			autoUpdateWorkflowLog.Print("Auto-update workflow deleted successfully")
		}
		return nil
	}

	seed := buildAutoUpdateSeed(opts.RepoSlug)
	cronSchedule, err := parser.ScatterSchedule("FUZZY:WEEKLY", seed)
	if err != nil {
		return fmt.Errorf("failed to scatter FUZZY:WEEKLY schedule for auto-update workflow: %w", err)
	}
	autoUpdateWorkflowLog.Printf("Scattered FUZZY:WEEKLY to %q for seed %q", cronSchedule, seed)

	content := buildAutoUpdateWorkflowYAML(cronSchedule)

	autoUpdateWorkflowLog.Printf("Writing auto-update workflow to %s", outputFile)
	if err := fileutil.EnsureParentDir(outputFile, constants.DirPermPublic); err != nil {
		return fmt.Errorf("failed to create auto-update workflow directory: %w", err)
	}
	if err := os.WriteFile(outputFile, []byte(content), constants.FilePermPublic); err != nil {
		return fmt.Errorf("failed to write auto-update workflow: %w", err)
	}

	autoUpdateWorkflowLog.Print("Auto-update workflow generated successfully")
	return nil
}

// buildAutoUpdateSeed returns the deterministic seed string used to scatter the
// FUZZY:WEEKLY cron schedule. It combines the repo slug with the fixed workflow
// identifier so that repositories scatter to distinct time slots.
func buildAutoUpdateSeed(repoSlug string) string {
	if repoSlug != "" {
		return repoSlug + "/" + autoUpdateWorkflowIdentifier
	}
	return autoUpdateWorkflowIdentifier
}

// buildAutoUpdateWorkflowYAML generates the YAML content for agentic-update.yml.
func buildAutoUpdateWorkflowYAML(cronSchedule string) string {
	customInstructions := `Alternative regeneration methods:
  make recompile

Or use the gh-aw CLI directly:
  ./gh-aw compile --validate --verbose

The workflow is generated when maintenance.auto_updates is set to true in aw.json.
The weekly schedule is deterministically scattered based on the repository slug.`

	header := GenerateWorkflowHeader("", "pkg/workflow/auto_update_workflow.go", customInstructions)

	return header + `name: Agentic Update

on:
  schedule:
    - cron: "` + cronSchedule + `"  # Weekly (auto-update)
  workflow_dispatch:

permissions:
  actions: write
  contents: write
  pull-requests: write

jobs:
  update:
    uses: ./.github/workflows/agentics-maintenance.yml
    with:
      operation: update
`
}
