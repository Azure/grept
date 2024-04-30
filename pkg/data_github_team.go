package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/shurcooL/githubv4"
)

var _ Data = &GitHubTeamDatasource{}

type GitHubTeamDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner          string   `hcl:"owner"`
	Slug           string   `hcl:"slug"`
	MembershipType string   `hcl:"membership_type,optional" default:"all" validate:"oneof=all immediate"`
	TeamName       string   `attribute:"team_name"`
	Description    string   `attribute:"description"`
	Privacy        string   `attribute:"privacy"`
	Permission     string   `attribute:"permission"`
	NodeId         string   `attribute:"node_id"`
	TeamId         int64    `attribute:"team_id"`
	Members        []string `attribute:"members"`
}

func (g *GitHubTeamDatasource) Type() string {
	return "github_team"
}

func (g *GitHubTeamDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	org, _, err := client.Organizations.Get(g.Context(), g.Owner)
	if err != nil {
		return fmt.Errorf("cannot read org info for %s, %s must be an organization", g.Owner, g.Owner)
	}
	team, resp, err := client.Teams.GetTeamBySlug(g.Context(), g.Owner, g.Slug)
	if err != nil {
		if resp.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("cannot get team by slug: %+v", err)
	}
	g.Description = value(team.Description)
	g.NodeId = value(team.NodeID)
	g.Privacy = value(team.Privacy)
	g.Permission = value(team.Permission)
	g.TeamId = value(team.ID)
	g.TeamName = value(team.Name)
	members, err := g.getTeamMembers(client, org, team)
	if err != nil {
		return fmt.Errorf("error when read team member for team %s: %+v", g.Slug, err)
	}
	g.Members = members
	return nil
}

func (g *GitHubTeamDatasource) getTeamMembers(client *githubclient.Client, org *github.Organization, team *github.Team) ([]string, error) {
	var members []string
	options := github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	if g.MembershipType == "all" {
		for {
			member, resp, err := client.Teams.ListTeamMembersByID(g.Context(), value(org.ID), team.GetID(), &options)
			if err != nil {
				return nil, err
			}

			for _, v := range member {
				members = append(members, v.GetLogin())
			}

			if resp.NextPage == 0 {
				break
			}
			options.Page = resp.NextPage
		}
		return members, nil
	}
	type member struct {
		Login string
	}
	var query struct {
		Organization struct {
			Team struct {
				Members struct {
					Nodes    []member
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"members(first:100,after:$memberCursor,membership:IMMEDIATE)"`
			} `graphql:"team(slug:$slug)"`
		} `graphql:"organization(login:$owner)"`
	}
	variables := map[string]interface{}{
		"owner":        githubv4.String(g.Owner),
		"slug":         githubv4.String(g.Slug),
		"memberCursor": (*githubv4.String)(nil),
	}
	gClient := client.GraphQLClient
	for {
		nameErr := gClient.Query(g.Context(), &query, variables)
		if nameErr != nil {
			return nil, nameErr
		}
		for _, v := range query.Organization.Team.Members.Nodes {
			members = append(members, v.Login)
		}
		if query.Organization.Team.Members.PageInfo.HasNextPage {
			variables["memberCursor"] = query.Organization.Team.Members.PageInfo.EndCursor
		} else {
			break
		}
	}
	return members, nil
}
