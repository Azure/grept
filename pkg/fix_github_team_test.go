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
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("to run this test you must set env GITHUB_TOKEN first")
	}
	owner := readEssentialEnv(t, "GITHUB_TEAM_FIX_INTEGRATION_TEST_OWNER")
	expectedTeam := readEssentialEnv(t, "GITHUB_TEAM_FIX_INTEGRATION_TEST_EXPECTED_TEAM")
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)

	defer func() {
		_, _ = client.Teams().DeleteTeamBySlug(context.Background(), owner, expectedTeam)
	}()
	sut := &pkg.GitHubTeamFix{
		Owner:                   owner,
		TeamName:                expectedTeam,
		ParentTeamId:            -1,
		CreateDefaultMaintainer: false,
	}
	err = sut.Apply()
	require.NoError(t, err)
	team, _, err := client.Teams().GetTeamBySlug(context.Background(), owner, expectedTeam)
	require.NoError(t, err)
	assert.Equal(t, expectedTeam, *team.Name)
}
