package pkg_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubRepositoryTeamFix(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("run integration test only under linux to avoid parallel testing issue")
	}
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	repoName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_REPO_NAME")
	altTeamName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_ALT_TEAM")
	expectedTeam := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_TEAM")
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	org, _, err := client.Organizations.Get(context.Background(), owner)
	require.NoError(t, err)
	altTeam, _, err := client.Teams.GetTeamBySlug(context.Background(), owner, altTeamName)
	require.NoError(t, err)
	defer func() {
		_, _ = client.Teams.RemoveTeamRepoByID(context.Background(), *org.ID, *altTeam.ID, owner, repoName)
	}()
	sut := &pkg.GitHubTeamRepositoryFix{
		Owner:    owner,
		RepoName: repoName,
		Teams: []pkg.TeamRepositoryBinding{
			{
				TeamSlug:   expectedTeam,
				Permission: "admin",
			},
			{
				TeamSlug:   altTeamName,
				Permission: "pull",
			},
		},
	}
	err = sut.Apply()
	require.NoError(t, err)
	teams, _, err := client.Repositories.ListTeams(context.Background(), owner, repoName, &github.ListOptions{PerPage: 100})
	require.NoError(t, err)
	assert.Len(t, teams, 2)
	parsedTeams := make(map[string]string)
	for _, team := range teams {
		parsedTeams[*team.Name] = *team.Permission
	}
	assert.Equal(t, map[string]string{
		expectedTeam: "admin",
		altTeamName:  "pull",
	}, parsedTeams)
}
