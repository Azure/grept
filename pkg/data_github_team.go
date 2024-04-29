package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
)

var _ Data = &GitHubTeamDatasource{}

type GitHubTeamDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner       string `hcl:"owner"`
	Slug        string `hcl:"slug"`
	TeamName    string `attribute:"team_name"`
	Description string `attribute:"description"`
	Privacy     string `attribute:"privacy"`
	Permission  string `attribute:"permission"`
	NodeId      string `attribute:"node_id"`
	TeamId      int64  `attribute:"team_id"`
}

func (g *GitHubTeamDatasource) Type() string {
	return "github_team"
}

func (g *GitHubTeamDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	team, _, err := client.Teams().GetTeamBySlug(g.Context(), g.Owner, g.Slug)
	if err != nil {
		return fmt.Errorf("cannot get team by slug: %+v", err)
	}
	g.Description = value(team.Description)
	g.NodeId = value(team.NodeID)
	g.Privacy = value(team.Privacy)
	g.Permission = value(team.Permission)
	g.TeamId = value(team.ID)
	g.TeamName = value(team.Name)
	return nil
}
