package pkg_test

import (
	"context"
	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGitHubTeamMembersFix(t *testing.T) {
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	expectedTeam := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_TEAM")
	expectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_USER_NAME")
	client, err := githubclient.GetGithubClient()
	org, _, err := client.Organizations.Get(context.Background(), owner)
	require.NoError(t, err)
	team, _, err := client.Teams.GetTeamBySlug(context.Background(), owner, expectedTeam)
	require.NoError(t, err)
	defer func() {
		members, _, err := client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		})
		require.NoError(t, err)
		for _, member := range members {
			_, _ = client.Teams.RemoveTeamMembershipByID(context.Background(), *org.ID, *team.ID, *member.Login)
		}
		_, _, err = client.Teams.AddTeamMembershipBySlug(context.Background(), owner, expectedTeam, expectedUserName, &github.TeamAddTeamMembershipOptions{
			Role: "maintainer",
		})
		require.NoError(t, err)
	}()
	require.NoError(t, err)
	sut := &pkg.GitHubTeamMembersFix{
		Owner:    owner,
		TeamSlug: expectedTeam,
		Members:  []pkg.TeamMember{},
	}
	err = sut.Apply()
	require.NoError(t, err)
	members, _, err := client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Empty(t, members)
	sut.Members = []pkg.TeamMember{
		{
			UserName: expectedUserName,
			Role:     "maintainer",
		},
	}
	err = sut.Apply()
	require.NoError(t, err)
	members, _, err = client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, expectedUserName, *members[0].Login)
	membership, _, err := client.Teams.GetTeamMembershipBySlug(context.Background(), owner, *team.Slug, expectedUserName)
	require.NoError(t, err)
	assert.Equal(t, "maintainer", *membership.Role)
	sut = &pkg.GitHubTeamMembersFix{
		Owner:    owner,
		TeamSlug: expectedTeam,
		Members:  []pkg.TeamMember{},
	}
	err = sut.Apply()
	require.NoError(t, err)
	members, _, err = client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Empty(t, members)
}
