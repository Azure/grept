package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"net/url"
)

var _ Fix = &GitHubRepositoryEnvironmentsFix{}

type ReviewerForGitHubRepositoryEnvironmentsFix struct {
	TeamId *int64 `hcl:"team_id,optional"`
	UserId *int64 `hcl:"user_id,optional"`
}

type DeploymentBranchPolicyForGitHubRepositoryEnvironmentsFix struct {
	ProtectedBranches    bool `hcl:"protected_branches"`
	CustomBranchPolicies bool `hcl:"custom_branch_policies"`
}

type EnvironmentForGitHubRepositoryEnvironmentsFix struct {
	Name                   string                                                    `hcl:"name"`
	CanAdminsBypass        bool                                                      `hcl:"can_admins_bypass,optional" default:"true"`
	PreventSelfReview      bool                                                      `hcl:"prevent_self_review,optional" default:"false"`
	WaitTimer              *int                                                      `hcl:"wait_timer,optional" validate:"gte=0,lte=43200"`
	Reviewers              []ReviewerForGitHubRepositoryEnvironmentsFix              `hcl:"reviewer,optional" validate:"max=6"`
	DeploymentBranchPolicy *DeploymentBranchPolicyForGitHubRepositoryEnvironmentsFix `hcl:"deployment_branch_policy,optional"`
}

type GitHubRepositoryEnvironmentsFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner        string                                          `hcl:"owner"`
	RepoName     string                                          `hcl:"repo_name"`
	Environments []EnvironmentForGitHubRepositoryEnvironmentsFix `hcl:"environment"`
}

func (g *GitHubRepositoryEnvironmentsFix) Type() string {
	return "github_repository_environments"
}

func (g *GitHubRepositoryEnvironmentsFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	environments, err := listGitHubRepositoryEnvironments(client, g.Owner, g.RepoName)
	if err != nil {
		return fmt.Errorf("cannot list environments for %s/%s: %+v", g.Owner, g.RepoName, err)
	}
	existingEnvs := make(map[string]struct{})
	for _, environment := range environments {
		existingEnvs[environment.Name] = struct{}{}
	}
	for _, environment := range g.Environments {
		var reviewers []*github.EnvReviewers
		for _, r := range environment.Reviewers {
			reviewerType := "User"
			id := r.UserId
			if r.TeamId != nil {
				reviewerType = "Team"
				id = r.TeamId
			}
			reviewers = append(reviewers, &github.EnvReviewers{
				Type: github.String(reviewerType),
				ID:   id,
			})
		}

		_, _, err = client.Repositories.CreateUpdateEnvironment(g.Context(), g.Owner, g.RepoName, url.PathEscape(environment.Name), &github.CreateUpdateEnvironment{
			WaitTimer:       environment.WaitTimer,
			Reviewers:       reviewers,
			CanAdminsBypass: &environment.CanAdminsBypass,
			//TODO
			DeploymentBranchPolicy: nil,
			PreventSelfReview:      &environment.PreventSelfReview,
		})
		if err != nil {
			return fmt.Errorf("cannot create or update environment %s for %s/%s: %+v", environment.Name, g.Owner, g.RepoName, err)
		}
	}
	return nil
}
