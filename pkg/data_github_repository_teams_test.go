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

func TestGithubRepositoryTeamsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepositoriesClient := NewMockRepositoriesClient(ctrl)
	owner := "grept"
	name := "test"
	sut := &pkg.GitHubRepositoryTeamsDatasource{
		Owner:    owner,
		RepoName: name,
	}
	stub := gostub.Stub(&githubclient.GetGithubClient, func() (*githubclient.Client, error) {
		mockRepositoriesClient.EXPECT().ListTeams(gomock.Any(), gomock.Eq(owner), gomock.Eq(name), gomock.Any()).Return([]*github.Team{
			{
				Name:       p(name),
				Slug:       p("slug"),
				Permission: p("permission"),
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
	assert.Len(t, sut.Teams, 1)
	assert.Equal(t, name, sut.Teams[0].Name)
	assert.Equal(t, "slug", sut.Teams[0].Slug)
	assert.Equal(t, "permission", sut.Teams[0].Permission)
}

func TestGithubRepositoryTeamsRead_IntegrationTest(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("to run this test you must set env GITHUB_TOKEN first")
	}
	owner := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_OWNER")
	repoName := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_REPO_NAME")
	expectedTeam := readEssentialEnv(t, "GITHUB_REPOSITORY_TEAMS_DATASOURCE_INTEGRATION_TEST_EXPECTED_TEAM")

	sut := &pkg.GitHubRepositoryTeamsDatasource{
		Owner:    owner,
		RepoName: repoName,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	matched := false
	for _, team := range sut.Teams {
		if team.Name == expectedTeam {
			matched = true
			break
		}
	}
	assert.True(t, matched)
}

func readEssentialEnv(t *testing.T, envName string) string {
	r := os.Getenv(envName)
	if r == "" {
		t.Skipf("to run this test you must set env %s first", envName)
	}
	return r
}

func p[T any](s T) *T {
	return &s
}
