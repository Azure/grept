package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Fix = &GitHubTeamRepositoryFix{}

type TeamRepositoryBinding struct {
	TeamSlug   string `hcl:"team_slug"`
	Permission string `hcl:"permission,optional" default:"pull" validate:"oneof=pull triage push maintain admin"`
}

type GitHubTeamRepositoryFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner    string                  `hcl:"owner"`
	RepoName string                  `hcl:"repo_name"`
	Teams    []TeamRepositoryBinding `hcl:"team,optional"`
}

func (g *GitHubTeamRepositoryFix) Type() string {
	return "github_team_repository"
}

func (g *GitHubTeamRepositoryFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	teamClient := client.Teams()
	org, _, err := client.Organizations().Get(g.Context(), g.Owner)
	if err != nil {
		return fmt.Errorf("cannot read org info for %s, %s must be an organization", g.Owner, g.Owner)
	}
	var teamIds []int64
	for _, team := range g.Teams {
		teamId, err := g.checkTeamSlug(client, team.TeamSlug)
		if err != nil {
			return fmt.Errorf("cannot read team id for team %s", team.TeamSlug)
		}
		teamIds = append(teamIds, teamId)
	}
	teams, err := listTeamsForRepository(client, g.Owner, g.RepoName, g.Context())
	if err != nil {
		return fmt.Errorf("cannot read existing teams for %s/%s: %+v", g.Owner, g.RepoName, err)
	}
	for _, team := range teams {
		if _, err := teamClient.RemoveTeamRepoByID(g.Context(), *org.ID, *team.ID, g.Owner, g.RepoName); err != nil {
			return fmt.Errorf("error when remove team %s from repo %s: %+v", *team.Name, g.RepoName, err)
		}
	}
	for i, newTeamId := range teamIds {
		if _, err := teamClient.AddTeamRepoByID(g.Context(), *org.ID, newTeamId, g.Owner, g.RepoName, &github.TeamAddTeamRepoOptions{Permission: g.Teams[i].Permission}); err != nil {
			return fmt.Errorf("error when add team %s to repo %s/%s: %+v", g.Teams[i].TeamSlug, g.Owner, g.RepoName, err)
		}
	}
	return nil
}

func (g *GitHubTeamRepositoryFix) checkTeamSlug(client *githubclient.Client, slug string) (int64, error) {
	team, _, err := client.Teams().GetTeamBySlug(g.Context(), g.Owner, slug)
	if err != nil {
		return -1, err
	}
	return *team.ID, nil
}
