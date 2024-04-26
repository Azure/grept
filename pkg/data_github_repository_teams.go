package pkg

import (
	"context"
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Data = &GitHubRepositoryTeamsDatasource{}

type Team struct {
	Name       string `hcl:"name"`
	Slug       string `hcl:"slug"`
	Permission string `hcl:"permission"`
}

type GitHubRepositoryTeamsDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner    string `hcl:"owner"`
	RepoName string `hcl:"repo_name"`
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
	githubTeams, err := listTeamsForRepository(client, g.Owner, g.RepoName, g.Context())
	if err != nil {
		return err
	}
	var teams []Team
	for _, team := range githubTeams {
		teams = append(teams, Team{
			Name:       value(team.Name),
			Slug:       value(team.Slug),
			Permission: value(team.Permission),
		})
	}
	g.Teams = teams
	return nil
}

func listTeamsForRepository(client *githubclient.Client, owner, repoName string, ctx context.Context) ([]*github.Team, error) {
	opts := &github.ListOptions{PerPage: 100}
	var r []*github.Team
	for {
		teams, resp, err := client.Repositories().ListTeams(ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("cannot list teams for %s/%s: %s", owner, repoName, err.Error())
		}
		r = append(r, teams...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return r, nil
}
