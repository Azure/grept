package pkg_test

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubTeamFixApply_IntegrationTest(t *testing.T) {
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	newTeam := "testNewTeam"
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)

	defer func() {
		_, _ = client.Teams.DeleteTeamBySlug(context.Background(), owner, newTeam)
	}()
	sut := &pkg.GitHubTeamFix{
		Owner:                   owner,
		TeamName:                newTeam,
		ParentTeamId:            -1,
		CreateDefaultMaintainer: false,
	}
	err = sut.Apply()
	require.NoError(t, err)
	team, _, err := client.Teams.GetTeamBySlug(context.Background(), owner, newTeam)
	require.NoError(t, err)
	assert.Equal(t, newTeam, *team.Name)
}
