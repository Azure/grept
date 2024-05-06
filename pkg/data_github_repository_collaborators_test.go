package pkg_test

import (
	"github.com/Azure/grept/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"runtime"
	"testing"
)

func TestGithubRepositoryCollaboratorsRead_IntegrationTest(t *testing.T) {
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
	expectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_USER_NAME")

	sut := &pkg.GitHubRepositoryCollaboratorsDatasource{
		Owner:    owner,
		RepoName: repoName,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	matched := false
	for _, collaborator := range sut.Collaborators {
		if collaborator.Login == expectedUserName {
			matched = true
			break
		}
	}
	assert.True(t, matched)
}
