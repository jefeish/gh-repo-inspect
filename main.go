package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jefeish/gh-repo-inspect/utils"
	"github.com/spf13/cobra"
)

type RepoInfo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type GovernanceConfig struct {
	Repository       RepoInfo           `json:"repository"`
	Rulesets         []Ruleset          `json:"rulesets,omitempty"`
	RequiredChecks   []string           `json:"required_checks,omitempty"`
	Collaborators    []Collaborator     `json:"collaborators,omitempty"`
	Teams            []Team             `json:"teams,omitempty"`
	SecuritySettings SecuritySettings   `json:"security_settings"`
	RepoSettings     RepositorySettings `json:"repository_settings"`
	IssueLabels      []Label            `json:"issue_labels,omitempty"`
	Milestones       []Milestone        `json:"milestones,omitempty"`
}

type Ruleset struct {
	Name                           string   `json:"name"`
	Pattern                        string   `json:"pattern"`
	EnforceAdmins                  bool     `json:"enforce_admins"`
	RequiredStatusChecks           []string `json:"required_status_checks,omitempty"`
	RequiredPullRequestReviews     bool     `json:"required_pull_request_reviews"`
	RequiredApprovingReviewCount   int      `json:"required_approving_review_count"`
	DismissStaleReviews            bool     `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews        bool     `json:"require_code_owner_reviews"`
	RequiredLinearHistory          bool     `json:"required_linear_history"`
	AllowForcePushes               bool     `json:"allow_force_pushes"`
	AllowDeletions                 bool     `json:"allow_deletions"`
	RequiredConversationResolution bool     `json:"required_conversation_resolution"`
}

type Collaborator struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
	Type       string `json:"type"`
}

type Team struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Permission string `json:"permission"`
}

type SecuritySettings struct {
	VulnerabilityAlerts          bool `json:"vulnerability_alerts"`
	AutomatedSecurityFixes       bool `json:"automated_security_fixes"`
	SecretScanning               bool `json:"secret_scanning"`
	SecretScanningPushProtection bool `json:"secret_scanning_push_protection"`
	DependencyGraphEnabled       bool `json:"dependency_graph_enabled"`
}

type RepositorySettings struct {
	Private             bool   `json:"private"`
	Archived            bool   `json:"archived"`
	Disabled            bool   `json:"disabled"`
	DefaultBranch       string `json:"default_branch"`
	AllowMergeCommit    bool   `json:"allow_merge_commit"`
	AllowSquashMerge    bool   `json:"allow_squash_merge"`
	AllowRebaseMerge    bool   `json:"allow_rebase_merge"`
	AllowAutoMerge      bool   `json:"allow_auto_merge"`
	DeleteBranchOnMerge bool   `json:"delete_branch_on_merge"`
	HasIssues           bool   `json:"has_issues"`
	HasProjects         bool   `json:"has_projects"`
	HasWiki             bool   `json:"has_wiki"`
	HasDownloads        bool   `json:"has_downloads"`
}

type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

type Milestone struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	State       string `json:"state"`
	DueOn       string `json:"due_on,omitempty"`
}

var (
	outputFormat string
	verbose      bool
	sections     []string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "repo-inspect [owner/repo]",
		Short: "Discover repository governance configuration",
		Long: `gh repo-inspect is a read-only GitHub CLI extension for discovering 
repository governance configuration without making changes.

This tool inspects various aspects of repository governance including:
- Repository rulesets and branch protection
- Required status checks
- Collaborators and teams
- Security settings
- Repository configuration
- Issue labels and milestones`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInspect,
	}

	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json, yaml, table)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringSliceVarP(&sections, "sections", "s", []string{}, "Specific sections to inspect (rulesets, collaborators, teams, security, settings, labels, milestones)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInspect(cmd *cobra.Command, args []string) error {
	var repo string
	if len(args) == 0 {
		// Try to get repo from current directory
		currentRepo, err := getCurrentRepo()
		if err != nil {
			return fmt.Errorf("no repository specified and could not determine current repository: %v", err)
		}
		repo = currentRepo
	} else {
		repo = args[0]
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repository must be in format 'owner/repo'")
	}

	owner, repoName := parts[0], parts[1]

	if verbose {
		fmt.Fprintf(os.Stderr, "Inspecting repository: %s/%s\n", owner, repoName)
	}

	governance, err := inspectRepository(owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to inspect repository: %v", err)
	}

	return outputGovernance(governance, sections)
}

func getCurrentRepo() (string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}

	// This is a simplified approach - in a real implementation,
	// you might want to parse .git/config or use git commands
	response := struct {
		FullName string `json:"full_name"`
	}{}

	err = client.Get("user/repos", &response)
	if err != nil {
		return "", err
	}

	return response.FullName, nil
}

func inspectRepository(owner, repo string) (*GovernanceConfig, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}

	governance := &GovernanceConfig{
		Repository: RepoInfo{
			Owner: owner,
			Name:  repo,
		},
	}

	// Get repository basic information
	if err := getRepositorySettings(*client, owner, repo, governance); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to get repository settings: %v\n", err)
		}
	}

	// Get rulesets if requested or if no specific sections
	if shouldIncludeSection("rulesets") {
		if err := getRulesets(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get rulesets: %v\n", err)
			}
		}
	}

	// Get collaborators if requested or if no specific sections
	if shouldIncludeSection("collaborators") {
		if err := getCollaborators(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get collaborators: %v\n", err)
			}
		}
	}

	// Get teams if requested or if no specific sections
	if shouldIncludeSection("teams") {
		if err := getTeams(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get teams: %v\n", err)
			}
		}
	}

	// Get security settings if requested or if no specific sections
	if shouldIncludeSection("security") {
		if err := getSecuritySettings(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get security settings: %v\n", err)
			}
		}
	}

	// Get labels if requested or if no specific sections
	if shouldIncludeSection("labels") {
		if err := getLabels(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get labels: %v\n", err)
			}
		}
	}

	// Get milestones if requested or if no specific sections
	if shouldIncludeSection("milestones") {
		if err := getMilestones(*client, owner, repo, governance); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get milestones: %v\n", err)
			}
		}
	}
		Collaborators: []Collaborator{
			{
				Login:      "maintainer1",
				Permission: "admin",
				Type:       "User",
			},
			{
				Login:      "developer1",
				Permission: "write",
				Type:       "User",
			},
		},
		Teams: []Team{
			{
				Name:       "Core Team",
				Slug:       "core-team",
				Permission: "admin",
			},
			{
				Name:       "Contributors",
				Slug:       "contributors",
				Permission: "write",
			},
			{
				Name:       "Reviewers",
				Slug:       "reviewers",
				Permission: "triage",
			},
		},
		SecuritySettings: SecuritySettings{
			VulnerabilityAlerts:          true,
			AutomatedSecurityFixes:       true,
			SecretScanning:               true,
			SecretScanningPushProtection: true,
			DependencyGraphEnabled:       true,
		},
		RepoSettings: RepositorySettings{
			Private:             false,
			Archived:            false,
			DefaultBranch:       "main",
			AllowMergeCommit:    true,
			AllowSquashMerge:    true,
			AllowRebaseMerge:    true,
			DeleteBranchOnMerge: true,
			HasIssues:           true,
			HasProjects:         false,
			HasWiki:             false,
		},
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Note: Using mock data for demonstration. In production, this would fetch real data from GitHub API.\n")
	}

	return governance, nil
}

func shouldIncludeSection(section string) bool {
	return utils.ShouldIncludeSection(sections, section)
}
