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

func TestGithubRepositoryCollaboratorsFix_IntegrationTest(t *testing.T) {
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
	expectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_EXPECTED_USER_NAME")
	altExpectedUserName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_ALT_EXPECTED_USER_NAME")

	defer func() {
		collaborators, _, err := client.Repositories.ListCollaborators(context.Background(), owner, repoName, &github.ListCollaboratorsOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		})
		require.NoError(t, err)
		for _, collaborator := range collaborators {
			invitation, _ := pkg.FindRepoInvitation(client, context.Background(), owner, repoName, *collaborator.Login)
			if invitation != nil {
				_, err = client.Repositories.DeleteInvitation(context.Background(), owner, repoName, invitation.GetID())
				require.NoError(t, err)
			}
			_, err = client.Repositories.RemoveCollaborator(context.Background(), owner, repoName, *collaborator.Login)
			require.NoError(t, err)
		}
		_, _, err = client.Repositories.AddCollaborator(context.Background(), owner, repoName, expectedUserName, &github.RepositoryAddCollaboratorOptions{
			Permission: "admin",
		})
		require.NoError(t, err)
	}()

	cases := []map[string]string{
		{},
		{
			expectedUserName: "admin",
		},
		{
			expectedUserName:    "admin",
			altExpectedUserName: "pull",
		},
		{
			expectedUserName: "admin",
		},
		{
			altExpectedUserName: "pull",
		},
		{},
	}
	for _, expectedCollaborators := range cases {
		sut := &pkg.GitHubRepositoryCollaboratorsFix{
			Owner:    owner,
			RepoName: repoName,
		}
		for name, permission := range expectedCollaborators {
			sut.Collaborators = append(sut.Collaborators, pkg.CollaboratorForRepositoryCollaboratorsFix{
				Name:       name,
				Permission: permission,
			})
		}
		err = sut.Apply()
		require.NoError(t, err)
		assertCollaborators(t, client, owner, repoName, expectedCollaborators)
	}
}

func assertCollaborators(t *testing.T, client *githubclient.Client, owner, repoName string, collaborators map[string]string) {
	result, _, err := client.Repositories.ListCollaborators(context.Background(), owner, repoName, &github.ListCollaboratorsOptions{
		Affiliation: "direct",
		ListOptions: github.ListOptions{PerPage: 100},
	})
	require.NoError(t, err)
	assert.Equal(t, len(collaborators), len(result))
	for _, user := range result {
		expectedPermission, ok := collaborators[*user.Login]
		require.True(t, ok)
		var matched = false
		for permission, enabled := range user.Permissions {
			if enabled && expectedPermission == permission {
				matched = true
			}
		}
		assert.True(t, matched)
	}
}
