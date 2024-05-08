package pkg

import (
	"context"
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Data = &GitHubRepositoryEnvironmentsDatasource{}

type EnvironmentForGitHubRepositoryEnvironmentsDatasource struct {
	Name   string `attribute:"name"`
	NodeId string `attribute:"node_id"`
}

type GitHubRepositoryEnvironmentsDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner        string                                                 `hcl:"owner"`
	RepoName     string                                                 `hcl:"repo_name"`
	Environments []EnvironmentForGitHubRepositoryEnvironmentsDatasource `attribute:"environments"`
}

func (g *GitHubRepositoryEnvironmentsDatasource) Type() string {
	return "github_repository_environments"
}

func (g *GitHubRepositoryEnvironmentsDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	var results []EnvironmentForGitHubRepositoryEnvironmentsDatasource
	var listOptions *github.EnvironmentListOptions
	for {
		environments, resp, err := client.Repositories.ListEnvironments(context.Background(), g.Owner, g.RepoName, listOptions)
		if err != nil {
			return err
		}
		if environments == nil {
			return nil
		}
		for _, environment := range environments.Environments {
			results = append(results, EnvironmentForGitHubRepositoryEnvironmentsDatasource{
				Name:   environment.GetName(),
				NodeId: environment.GetNodeID(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	g.Environments = results
	return nil
}
