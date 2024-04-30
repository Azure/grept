package pkg

import (
	"context"
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Fix = &GitHubTeamMembersFix{}

type TeamMember struct {
	UserName string `hcl:"username"`
	Role     string `hcl:"role,optional" validate:"oneof=member maintainer" default:"member"`
}

type GitHubTeamMembersFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner    string       `hcl:"owner"`
	TeamSlug string       `hcl:"team_slug"`
	Members  []TeamMember `hcl:"member,optional"`
}

func (g GitHubTeamMembersFix) Type() string {
	return "github_team_members"
}

func (g GitHubTeamMembersFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	org, _, err := client.Organizations.Get(g.Context(), g.Owner)
	if err != nil {
		return fmt.Errorf("cannot read org info for %s, %s must be an organization", g.Owner, g.Owner)
	}
	team, err := readTeam(client, g.Context(), g.Owner, g.TeamSlug)
	if err != nil {
		return err
	}
	expectedMembers := make(map[string]TeamMember)
	for _, member := range g.Members {
		expectedMembers[member.UserName] = member
	}
	for {
		opts := &github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		members, resp, err := client.Teams.ListTeamMembersByID(g.Context(), *org.ID, *team.ID, opts)
		if err != nil {
			return fmt.Errorf("cannot list members for %s/%s: %s", g.Owner, g.TeamSlug, err.Error())
		}
		for _, c := range members {
			expectedMembership, found := expectedMembers[*c.Login]
			if !found {
				_, err := client.Teams.RemoveTeamMembershipBySlug(g.Context(), g.Owner, g.TeamSlug, *c.Login)
				if err != nil {
					return fmt.Errorf("cannot remove membership %s from %s/%s", *c.Login, g.Owner, g.TeamSlug)
				}
				continue
			}
			membership, _, err := client.Teams.GetTeamMembershipBySlug(g.Context(), g.Owner, g.TeamSlug, expectedMembership.UserName)
			if err != nil {
				return fmt.Errorf("cannot get membership %s from %s/%s", *c.Login, g.Owner, g.TeamSlug)
			}
			if *membership.Role != expectedMembership.Role {
				_, err := client.Teams.RemoveTeamMembershipBySlug(g.Context(), g.Owner, g.TeamSlug, *c.Login)
				if err != nil {
					return fmt.Errorf("cannot remove membership %s from %s/%s", *c.Login, g.Owner, g.TeamSlug)
				}
				_, _, err = client.Teams.AddTeamMembershipBySlug(g.Context(), g.Owner, g.TeamSlug, expectedMembership.UserName, &github.TeamAddTeamMembershipOptions{Role: expectedMembership.Role})
				if err != nil {
					return fmt.Errorf("cannot add membership %s to %s/%s", *c.Name, g.Owner, g.TeamSlug)
				}
				delete(expectedMembers, expectedMembership.UserName)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	for name, member := range expectedMembers {
		_, _, err = client.Teams.AddTeamMembershipBySlug(g.Context(), g.Owner, g.TeamSlug, name, &github.TeamAddTeamMembershipOptions{Role: member.Role})
		if err != nil {
			return fmt.Errorf("cannot add membership %s to %s/%s", name, g.Owner, g.TeamSlug)
		}
	}
	return nil
}

func readTeam(client *githubclient.Client, ctx context.Context, owner, teamSlug string) (*github.Team, error) {
	var team *github.Team
	err := Do(func(attempt int) (retry bool, err error) {
		var resp *github.Response
		team, resp, err = client.Teams.GetTeamBySlug(ctx, owner, teamSlug)
		if err != nil {
			if resp.StatusCode == 404 {
				return attempt < 3, err
			}
			return false, fmt.Errorf("cannot get team by slug: %+v", err)
		}
		return false, nil
	})
	return team, err
}
