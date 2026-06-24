//go:build !integration

package parser

import (
	"strings"
	"testing"
)

func TestBuildCommitLookupAPIPath(t *testing.T) {
	t.Run("escapes refs containing slash", func(t *testing.T) {
		got := buildCommitLookupAPIPath("owner", "repo", "feature/github-agentic-workflows")
		want := "/repos/owner/repo/commits/feature%2Fgithub-agentic-workflows"
		if got != want {
			t.Fatalf("buildCommitLookupAPIPath() = %q, want %q", got, want)
		}
	})

	t.Run("keeps plain refs readable", func(t *testing.T) {
		got := buildCommitLookupAPIPath("owner", "repo", "main")
		want := "/repos/owner/repo/commits/main"
		if got != want {
			t.Fatalf("buildCommitLookupAPIPath() = %q, want %q", got, want)
		}
	})
}

func TestGitFallbackRequiresNonEmptyRef(t *testing.T) {
	t.Run("all files fallback validates ref", func(t *testing.T) {
		_, err := listDirAllFilesViaGitForHost("owner", "repo", "", "skills/demo", "")
		if err == nil {
			t.Fatal("expected error for empty ref")
		}
		if !strings.Contains(err.Error(), "non-empty ref") {
			t.Fatalf("expected non-empty ref error, got %q", err)
		}
	})

	t.Run("subdirs fallback validates ref", func(t *testing.T) {
		_, err := listDirSubdirsViaGitForHost("owner", "repo", "   ", "skills", "")
		if err == nil {
			t.Fatal("expected error for empty ref")
		}
		if !strings.Contains(err.Error(), "non-empty ref") {
			t.Fatalf("expected non-empty ref error, got %q", err)
		}
	})
}

func TestListContentsRecursivelyWithDepth_MaxDepthGuard(t *testing.T) {
	_, err := listContentsRecursivelyWithDepth(nil, "owner", "repo", "main", "skills/demo/deep", 11, 10)
	if err == nil {
		t.Fatal("expected depth limit error")
	}
	if !strings.Contains(err.Error(), "maximum skill directory recursion depth exceeded") {
		t.Fatalf("expected depth limit error, got %q", err)
	}
}

func TestGitRefValidation(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		expectError bool
	}{
		{
			name:        "valid branch name",
			ref:         "main",
			expectError: false,
		},
		{
			name:        "valid feature branch",
			ref:         "feature/safe",
			expectError: false,
		},
		{
			name:        "valid tag",
			ref:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "valid SHA",
			ref:         "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			expectError: false,
		},
		{
			name:        "invalid hyphen prefix",
			ref:         "-exploit",
			expectError: true,
		},
		{
			name:        "invalid double hyphen prefix",
			ref:         "--upload-pack=touch /tmp/pwned",
			expectError: true,
		},
		{
			name:        "invalid hyphen prefix with spaces",
			ref:         "  -invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGitRef(tt.ref)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for ref %q but got none", tt.ref)
				} else if !strings.Contains(err.Error(), "must not start with '-'") {
					t.Errorf("expected security error message, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for ref %q: %v", tt.ref, err)
				}
			}
		})
	}
}
