package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Fix = &GitHubTeamRepositoryFix{}

type GitHubTeamRepositoryFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner      string `hcl:"owner"`
	RepoName   string `hcl:"repo_name"`
	TeamSlug   string `hcl:"team_slug,optional" validate:"conflict_with=TeamId,at_least_one_of=TeamSlug TeamId"`
	TeamId     int    `hcl:"team_id,optional" validate:"conflict_with=TeamSlug,at_least_one_of=TeamSlug TeamId"`
	Permission string `hcl:"permission,optional" default:"pull" validate:"oneof=pull triage push maintain admin"`
}

func (g *GitHubTeamRepositoryFix) Type() string {
	return "github_team_repository"
}

func (g *GitHubTeamRepositoryFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	org, _, err := client.Organizations().Get(g.Context(), g.Owner)
	if err != nil {
		return fmt.Errorf("cannot read org info for %s, %s must be an organization", g.Owner, g.Owner)
	}
	var teamId = int64(g.TeamId)
	if teamId != 0 {
		if err = g.checkTeamId(client, *org.ID, teamId); err != nil {
			return err
		}
	}
	teamId, err = g.checkTeamSlug(client, g.TeamSlug)
	if err != nil {
		return err
	}
	githubTeams, err := listTeamsForRepository(client, g.Owner, g.RepoName, g.Context())
	if err != nil {
		return err
	}
	for _, team := range githubTeams {
		if *team.ID == teamId {
			return nil
		}
	}

	_, err = client.Teams().AddTeamRepoByID(g.Context(), *org.ID, teamId, g.Owner, g.RepoName, &github.TeamAddTeamRepoOptions{
		Permission: g.Permission,
	})
	return err
}

func (g *GitHubTeamRepositoryFix) checkTeamSlug(client *githubclient.Client, slug string) (int64, error) {
	team, _, err := client.Teams().GetTeamBySlug(g.Context(), g.Owner, slug)
	if err != nil {
		return -1, err
	}
	return *team.ID, nil
}

func (g *GitHubTeamRepositoryFix) checkTeamId(client *githubclient.Client, orgId, teamId int64) error {
	_, _, err := client.Teams().GetTeamByID(g.Context(), orgId, teamId)
	return err
}
