package pkg_test

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubTeamMembersFix(t *testing.T) {
	testEnv := githubTeamMemberTestPrepare(t)
	defer testEnv.cleaner()
	owner := testEnv.owner
	expectedTeam := testEnv.expectedTeam
	expectedUserName := testEnv.expectedUserName
	org := testEnv.org
	team := testEnv.team
	sut := &pkg.GitHubTeamMembersFix{
		Owner:             owner,
		TeamSlug:          expectedTeam,
		PruneExtraMembers: true,
		Members:           []pkg.TeamMember{},
	}
	err := sut.Apply()
	require.NoError(t, err)
	client, err := githubclient.GetGithubClient()
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
		Owner:             owner,
		TeamSlug:          expectedTeam,
		PruneExtraMembers: true,
		Members:           []pkg.TeamMember{},
	}
	err = sut.Apply()
	require.NoError(t, err)
	members, _, err = client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Empty(t, members)
}

func TestGitHubTeamMembersFix_PruneExtraMembersFalseShouldNotRemoveAnyMembers(t *testing.T) {
	testEnv := githubTeamMemberTestPrepare(t)
	defer testEnv.cleaner()
	owner := testEnv.owner
	expectedTeam := testEnv.expectedTeam
	expectedUserName := testEnv.expectedUserName
	org := testEnv.org
	team := testEnv.team
	sut := &pkg.GitHubTeamMembersFix{
		Owner:             owner,
		TeamSlug:          expectedTeam,
		PruneExtraMembers: false,
		Members:           []pkg.TeamMember{},
	}
	err := sut.Apply()
	require.NoError(t, err)
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	members, _, err := client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, members[0].GetLogin(), expectedUserName)
}

func TestGitHubTeamMembersFix_PruneExtraMembersFalseShouldAddNonExistMembers(t *testing.T) {
	testEnv := githubTeamMemberTestPrepare(t)
	defer testEnv.cleaner()
	owner := testEnv.owner
	expectedTeam := testEnv.expectedTeam
	expectedUserName := testEnv.expectedUserName
	org := testEnv.org
	team := testEnv.team
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	members, _, err := client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	for _, member := range members {
		_, err = client.Teams.RemoveTeamMembershipBySlug(context.TODO(), org.GetLogin(), expectedTeam, member.GetLogin())
		require.NoError(t, err)
	}
	sut := &pkg.GitHubTeamMembersFix{
		Owner:             owner,
		TeamSlug:          expectedTeam,
		PruneExtraMembers: false,
		Members: []pkg.TeamMember{
			pkg.TeamMember{
				UserName: expectedUserName,
				Role:     "member",
			},
		},
	}
	err = sut.Apply()
	require.NoError(t, err)
	members, _, err = client.Teams.ListTeamMembersByID(context.Background(), *org.ID, *team.ID, &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, expectedUserName, members[0].GetLogin())
}

type githubTeamMemberTestEnv struct {
	owner            string
	expectedTeam     string
	expectedUserName string
	org              *github.Organization
	team             *github.Team
	cleaner          func()
}

func githubTeamMemberTestPrepare(t *testing.T) githubTeamMemberTestEnv {
	//if runtime.GOOS != "linux" {
	//	t.Skip("run integration test only under linux to avoid parallel testing issue")
	//}
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	expectedTeam := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_TEAM")
	expectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_USER_NAME")
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	org, _, err := client.Organizations.Get(context.Background(), owner)
	require.NoError(t, err)
	team, _, err := client.Teams.GetTeamBySlug(context.Background(), owner, expectedTeam)
	require.NoError(t, err)
	return githubTeamMemberTestEnv{
		owner:            owner,
		expectedTeam:     expectedTeam,
		expectedUserName: expectedUserName,
		org:              org,
		team:             team,
		cleaner: func() {
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
		},
	}
}
