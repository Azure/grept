package pkg_test

import (
	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"os"
	"testing"

	"github.com/prashantv/gostub"
)

func TestGithubRepositoryCollaboratorsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepositoriesClient := NewMockRepositoriesClient(ctrl)
	owner := "grept"
	name := "test"
	sut := &pkg.GitHubRepositoryCollaboratorsDatasource{
		Owner:    owner,
		RepoName: name,
	}
	stub := gostub.Stub(&githubclient.GetGithubClient, func() (*githubclient.Client, error) {
		mockRepositoriesClient.EXPECT().ListCollaborators(gomock.Any(), gomock.Eq(owner), gomock.Eq(name), gomock.Any()).Return([]*github.User{
			&github.User{
				Name: p("John Doe"),
			},
		}, &github.Response{}, nil).Times(1)
		return &githubclient.Client{
			Repositories: func() githubclient.RepositoriesClient {
				return mockRepositoriesClient
			},
		}, nil
	})
	defer stub.Reset()
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.Len(t, sut.Collaborators, 1)
	assert.Equal(t, "John Doe", sut.Collaborators[0].Name)
}

func TestGithubRepositoryCollaboratorsRead_IntegrationTest(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("to run this test you must set env GITHUB_TOKEN first")
	}
	owner := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_OWNER")
	repoName := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_REPO_NAME")
	expectedUserName := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_EXPECTED_USER_NAME")

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
