package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jefeish/gh-repo-inspect/utils"
	"gopkg.in/yaml.v3"
)

func outputGovernance(governance *GovernanceConfig, sectionsFilter []string) error {
	switch strings.ToLower(outputFormat) {
	case "json":
		return outputJSON(governance)
	case "yaml", "yml":
		return outputYAML(governance)
	case "table":
		return outputTable(governance, sectionsFilter)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func outputJSON(governance *GovernanceConfig) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(governance)
}

func outputYAML(governance *GovernanceConfig) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(governance)
}

func outputTable(governance *GovernanceConfig, sectionsFilter []string) error {
	fmt.Printf("Repository Governance Report\n")
	fmt.Printf("═══════════════════════════\n\n")

	// Repository Information
	fmt.Printf("📁 Repository: %s/%s\n\n", governance.Repository.Owner, governance.Repository.Name)

	// Repository Settings
	if shouldIncludeSectionOutput("settings", sectionsFilter) {
		fmt.Printf("⚙️  Repository Settings\n")
		fmt.Printf("├─ Private: %s\n", boolToIcon(governance.RepoSettings.Private))
		fmt.Printf("├─ Archived: %s\n", boolToIcon(governance.RepoSettings.Archived))
		fmt.Printf("├─ Default Branch: %s\n", governance.RepoSettings.DefaultBranch)
		fmt.Printf("├─ Issues: %s\n", boolToIcon(governance.RepoSettings.HasIssues))
		fmt.Printf("├─ Projects: %s\n", boolToIcon(governance.RepoSettings.HasProjects))
		fmt.Printf("├─ Wiki: %s\n", boolToIcon(governance.RepoSettings.HasWiki))
		fmt.Printf("├─ Allow Merge Commit: %s\n", boolToIcon(governance.RepoSettings.AllowMergeCommit))
		fmt.Printf("├─ Allow Squash Merge: %s\n", boolToIcon(governance.RepoSettings.AllowSquashMerge))
		fmt.Printf("├─ Allow Rebase Merge: %s\n", boolToIcon(governance.RepoSettings.AllowRebaseMerge))
		fmt.Printf("└─ Delete Branch on Merge: %s\n\n", boolToIcon(governance.RepoSettings.DeleteBranchOnMerge))
	}

	// Security Settings
	if shouldIncludeSectionOutput("security", sectionsFilter) {
		fmt.Printf("🔒 Security Settings\n")
		fmt.Printf("├─ Vulnerability Alerts: %s\n", boolToIcon(governance.SecuritySettings.VulnerabilityAlerts))
		fmt.Printf("├─ Automated Security Fixes: %s\n", boolToIcon(governance.SecuritySettings.AutomatedSecurityFixes))
		fmt.Printf("├─ Secret Scanning: %s\n", boolToIcon(governance.SecuritySettings.SecretScanning))
		fmt.Printf("├─ Secret Scanning Push Protection: %s\n", boolToIcon(governance.SecuritySettings.SecretScanningPushProtection))
		fmt.Printf("└─ Dependency Graph: %s\n\n", boolToIcon(governance.SecuritySettings.DependencyGraphEnabled))
	}

	// Repository Rulesets
	if len(governance.Rulesets) > 0 && shouldIncludeSectionOutput("rulesets", sectionsFilter) {
		fmt.Printf("📜 Repository Rulesets\n")
		for i, ruleset := range governance.Rulesets {
			prefix := "├─"
			if i == len(governance.Rulesets)-1 {
				prefix = "└─"
			}
			fmt.Printf("%s %s (Pattern: %s)\n", prefix, ruleset.Name, ruleset.Pattern)

			// Show main settings
			fmt.Printf("   ├─ Enforce Admins: %s\n", boolToIcon(ruleset.EnforceAdmins))
			fmt.Printf("   ├─ Require PR Reviews: %s\n", boolToIcon(ruleset.RequiredPullRequestReviews))
			if ruleset.RequiredPullRequestReviews {
				fmt.Printf("   │  ├─ Required Approving Reviews: %d\n", ruleset.RequiredApprovingReviewCount)
				fmt.Printf("   │  ├─ Dismiss Stale Reviews: %s\n", boolToIcon(ruleset.DismissStaleReviews))
				fmt.Printf("   │  └─ Require Code Owner Reviews: %s\n", boolToIcon(ruleset.RequireCodeOwnerReviews))
			}

			// Show branch protection settings
			fmt.Printf("   ├─ Required Linear History: %s\n", boolToIcon(ruleset.RequiredLinearHistory))
			fmt.Printf("   ├─ Allow Force Pushes: %s\n", boolToIcon(ruleset.AllowForcePushes))
			fmt.Printf("   ├─ Allow Deletions: %s\n", boolToIcon(ruleset.AllowDeletions))
			fmt.Printf("   ├─ Require Conversation Resolution: %s\n", boolToIcon(ruleset.RequiredConversationResolution))

			// Show required status checks
			if len(ruleset.RequiredStatusChecks) > 0 {
				fmt.Printf("   └─ Required Status Checks:\n")
				for j, check := range ruleset.RequiredStatusChecks {
					checkPrefix := "├─"
					if j == len(ruleset.RequiredStatusChecks)-1 {
						checkPrefix = "└─"
					}
					fmt.Printf("      %s %s\n", checkPrefix, check)
				}
			} else {
				fmt.Printf("   └─ Required Status Checks: None\n")
			}

			// Add spacing between rulesets except for the last one
			if i < len(governance.Rulesets)-1 {
				fmt.Printf("   \n")
			}
		}
		fmt.Println()
	}

	// Collaborators
	if len(governance.Collaborators) > 0 && shouldIncludeSectionOutput("collaborators", sectionsFilter) {
		fmt.Printf("👥 Collaborators (%d)\n", len(governance.Collaborators))
		for i, collab := range governance.Collaborators {
			prefix := "├─"
			if i == len(governance.Collaborators)-1 {
				prefix = "└─"
			}
			fmt.Printf("%s %s (%s) - %s\n", prefix, collab.Login, collab.Type, permissionToIcon(collab.Permission))
		}
		fmt.Println()
	}

	// Teams
	if len(governance.Teams) > 0 && shouldIncludeSectionOutput("teams", sectionsFilter) {
		fmt.Printf("Teams (%d)\n", len(governance.Teams))
		for i, team := range governance.Teams {
			prefix := "├─"
			if i == len(governance.Teams)-1 {
				prefix = "└─"
			}
			fmt.Printf("%s %s (@%s) - %s\n", prefix, team.Name, team.Slug, permissionToIcon(team.Permission))
		}
		fmt.Println()
	}

	// Labels
	if len(governance.IssueLabels) > 0 && shouldIncludeSectionOutput("labels", sectionsFilter) {
		fmt.Printf("🏷️  Labels (%d)\n", len(governance.IssueLabels))
		for i, label := range governance.IssueLabels {
			prefix := "├─"
			if i == len(governance.IssueLabels)-1 {
				prefix = "└─"
			}
			description := ""
			if label.Description != "" {
				description = fmt.Sprintf(" (%s)", label.Description)
			}
			fmt.Printf("%s %s #%s%s\n", prefix, label.Name, label.Color, description)
		}
		fmt.Println()
	}

	// Milestones
	if len(governance.Milestones) > 0 && shouldIncludeSectionOutput("milestones", sectionsFilter) {
		fmt.Printf("🎯 Milestones (%d)\n", len(governance.Milestones))
		for i, milestone := range governance.Milestones {
			prefix := "├─"
			if i == len(governance.Milestones)-1 {
				prefix = "└─"
			}
			state := "🟢"
			if milestone.State == "closed" {
				state = "🔴"
			}
			dueDate := ""
			if milestone.DueOn != "" {
				dueDate = fmt.Sprintf(" (Due: %s)", milestone.DueOn)
			}
			fmt.Printf("%s %s %s%s\n", prefix, state, milestone.Title, dueDate)
			if milestone.Description != "" {
				fmt.Printf("   %s\n", milestone.Description)
			}
		}
		fmt.Println()
	}

	return nil
}

// shouldIncludeSectionOutput determines if a section should be included in output
func shouldIncludeSectionOutput(section string, sectionsFilter []string) bool {
	return utils.ShouldIncludeSection(sectionsFilter, section)
}

func boolToIcon(b bool) string {
	return utils.BoolToIcon(b)
}

func permissionToIcon(permission string) string {
	return utils.PermissionToIcon(permission)
}
