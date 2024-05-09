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

func TestGithubRepositoryEnvironmentsFix_IntegrationTest(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("run integration test only under linux to avoid parallel testing issue")
	}
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	repoName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_REPO_NAME")
	env := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_ENVIRONMENT")
	altEnv := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_ALT_ENVIRONMENT")
	expectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_USER_NAME")

	user, _, err := client.Users.Get(context.Background(), expectedUserName)
	require.NoError(t, err)
	defer func() {
		response, _, err := client.Repositories.ListEnvironments(context.Background(), owner, repoName, &github.EnvironmentListOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		})
		require.NoError(t, err)
		for _, environment := range response.Environments {
			_, err := client.Repositories.DeleteEnvironment(context.Background(), owner, repoName, *environment.Name)
			require.NoError(t, err)
		}
		_, _, err = client.Repositories.CreateUpdateEnvironment(context.Background(), owner, repoName, env, &github.CreateUpdateEnvironment{
			Reviewers: []*github.EnvReviewers{
				{
					Type: github.String("User"),
					ID:   user.ID,
				},
			},
		})
		require.NoError(t, err)
	}()

	cases := [][]string{
		{
			env,
		},
		{
			env,
			altEnv,
		},
		{
			altEnv,
		},
		{
			env,
		},
		{},
	}
	for _, expectedEnvs := range cases {
		sut := &pkg.GitHubRepositoryEnvironmentsFix{
			Owner:    owner,
			RepoName: repoName,
		}
		for _, expectedEnv := range expectedEnvs {
			sut.Environments = append(sut.Environments, pkg.EnvironmentForGitHubRepositoryEnvironmentsFix{
				Name:              expectedEnv,
				CanAdminsBypass:   false,
				PreventSelfReview: true,
				WaitTimer:         github.Int(60),
				Reviewers: []pkg.ReviewerForGitHubRepositoryEnvironmentsFix{
					{
						UserId: user.ID,
					},
				},
				DeploymentBranchPolicy: &pkg.DeploymentBranchPolicyForGitHubRepositoryEnvironmentsFix{
					ProtectedBranches:    true,
					CustomBranchPolicies: false,
				},
			})
		}
		err = sut.Apply()
		require.NoError(t, err)
		assertRepositoryEnvironments(t, client, owner, repoName, expectedEnvs, user.ID)
	}
}

func assertRepositoryEnvironments(t *testing.T, client *githubclient.Client, owner, repoName string, envs []string, reviewerId *int64) {
	response, _, err := client.Repositories.ListEnvironments(context.Background(), owner, repoName, &github.EnvironmentListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	expectedEnvs := make(map[string]struct{})
	for _, env := range envs {
		expectedEnvs[env] = struct{}{}
	}
	assert.Equal(t, len(envs), len(response.Environments))
	for _, environment := range response.Environments {
		_, exist := expectedEnvs[*environment.Name]
		assert.True(t, exist)
	}
}
