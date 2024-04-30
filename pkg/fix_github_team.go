package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	githubgraphql "github.com/shurcooL/githubv4"
)

var _ Fix = &GitHubTeamFix{}

type GitHubTeamFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner                   string `hcl:"owner"`
	TeamName                string `hcl:"team_name"`
	Description             string `hcl:"description,optional"`
	Privacy                 string `hcl:"privacy,optional" default:"secret" validate:"oneof=secret closed"`
	ParentTeamId            int64  `hcl:"parent_team_id,optional" default:"-1"`
	LdapDistinguishedName   string `hcl:"ldap_dn,optional"`
	CreateDefaultMaintainer bool   `hcl:"create_default_maintainer,optional" default:"false"`
}

func (g *GitHubTeamFix) Type() string {
	return "github_team"
}

func (g *GitHubTeamFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	_, _, err = client.Organizations.Get(g.Context(), g.Owner)
	if err != nil {
		return fmt.Errorf("cannot read org info for %s, %s must be an organization", g.Owner, g.Owner)
	}
	newTeam := github.NewTeam{
		Name:        g.TeamName,
		Description: &g.Description,
		Privacy:     &g.Privacy,
	}
	if g.ParentTeamId != -1 {
		newTeam.ParentTeamID = &g.ParentTeamId
	}
	if g.LdapDistinguishedName != "" {
		newTeam.LDAPDN = &g.LdapDistinguishedName
	}
	teamClient := client.Teams
	githubTeam, _, err := teamClient.CreateTeam(g.Context(), g.Owner, newTeam)
	if err != nil {
		return fmt.Errorf("cannot create team: %+v", err)
	}
	/*  Notes from github Terraform provider:
	When using a GitHub App for authentication, `members:write` permissions on the App are needed.

	However, when using a GitHub App, CreateTeam will not correctly nest the team under the parent,
	if the parent team was created by someone else than the GitHub App. In that case, the response
	object will contain a `nil` parent object.

	This can be resolved by using an additional call to EditTeamByID. This will be able to set the
	parent team correctly when using a GitHub App with `members:write` permissions.

	Note that this is best-effort: when running this with a PAT that does not have admin permissions
	on the parent team, the operation might still fail to set the parent team.
	*/
	if newTeam.ParentTeamID != nil && githubTeam.Parent == nil {
		_, _, err := teamClient.EditTeamByID(g.Context(),
			*githubTeam.Organization.ID,
			*githubTeam.ID,
			newTeam,
			false)

		if err != nil {
			return fmt.Errorf("cannot edit team by id: %+v", err)
		}
	}
	if !g.CreateDefaultMaintainer {
		if err = g.removeDefaultMaintainer(client, *githubTeam.Slug); err != nil {
			return fmt.Errorf("error when remove default maintainer: %+v", err)
		}
	}
	return nil
}

func (g *GitHubTeamFix) removeDefaultMaintainer(client *githubclient.Client, teamSlug string) error {
	type User struct {
		Login githubgraphql.String
	}

	var query struct {
		Organization struct {
			Team struct {
				Members struct {
					Nodes []User
				}
			} `graphql:"team(slug:$slug)"`
		} `graphql:"organization(login:$login)"`
	}
	variables := map[string]interface{}{
		"slug":  githubgraphql.String(teamSlug),
		"login": githubgraphql.String(g.Owner),
	}

	err := client.GraphQLClient.Query(g.Context(), &query, variables)
	if err != nil {
		return err
	}

	for _, user := range query.Organization.Team.Members.Nodes {
		_, err := client.Teams.RemoveTeamMembershipBySlug(g.Context(), g.Owner, teamSlug, string(user.Login))
		if err != nil {
			return err
		}
	}

	return nil
}
