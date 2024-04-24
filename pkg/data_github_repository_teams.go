package pkg

import (
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/palantir/go-githubapp/githubapp"
)

var _ Data = &GitHubRepositoryTeamsDatasource{}

type Config struct {
	Github githubapp.Config `yaml:"github"`
}

type Team struct {
	Name       string `hcl:"name" json:"name"`
	Slug       string `hcl:"slug" json:"slug"`
	Permission string `hcl:"permission" json:"permission"`
}

type GitHubRepositoryTeamsDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner    string `hcl:"owner" json:"owner"`
	RepoName string `hcl:"repo_name" json:"repo_name"`
	Teams    []Team `attribute:"teams"`
}

func (g *GitHubRepositoryTeamsDatasource) Type() string {
	return "github_repository_teams"
}

func (g *GitHubRepositoryTeamsDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	opts := &github.ListOptions{PerPage: 100}
	for {
		teams, resp, err := client.Repositories().ListTeams(g.Context(), g.Owner, g.RepoName, opts)
		if err != nil {
			return fmt.Errorf("cannot list teams for %s/%s: %s", g.Owner, g.RepoName, err.Error())
		}
		for _, team := range teams {
			g.Teams = append(g.Teams, Team{
				Name:       *team.Name,
				Slug:       *team.Slug,
				Permission: *team.Permission,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}
