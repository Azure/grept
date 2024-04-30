package pkg_test

import (
	"github.com/Azure/grept/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGithubTeamRead_IntegrationTest(t *testing.T) {
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	expectedTeam := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_TEAM")

	sut := &pkg.GitHubTeamDatasource{
		Owner: owner,
		Slug:  expectedTeam,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.Equal(t, expectedTeam, sut.TeamName)
}
