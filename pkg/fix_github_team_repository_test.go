package pkg_test

import (
	"context"
	"github.com/prashantv/gostub"
	"go.uber.org/mock/gomock"
	"os"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/ahmetb/go-linq/v3"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubRepositoryTeamFix_TeamAlreadyExistShouldNotCallAddTeamRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOrgClient := NewMockOrganizationsClient(ctrl)
	mockTeamsClients := NewMockTeamsClient(ctrl)
	mockRepositoriesClient := NewMockRepositoriesClient(ctrl)
	stub := gostub.Stub(&githubclient.GetGithubClient, func() (*githubclient.Client, error) {
		return &githubclient.Client{
			Repositories: func() githubclient.RepositoriesClient {
				return mockRepositoriesClient
			},
			Teams: func() githubclient.TeamsClient {
				return mockTeamsClients
			},
			Organizations: func() githubclient.OrganizationsClient {
				return mockOrgClient
			},
		}, nil
	})
	defer stub.Reset()
	owner := "owner"
	repoName := "repo"
	teamSlug := "team"
	teamId := int64(123)
	mockOrgClient.EXPECT().Get(gomock.Any(), gomock.Eq(owner)).Return(&github.Organization{}, nil, nil)
	mockTeamsClients.EXPECT().GetTeamBySlug(gomock.Any(), gomock.Eq(owner), gomock.Eq(teamSlug)).Return(&github.Team{
		ID: p(teamId),
	}, &github.Response{}, nil)
	mockRepositoriesClient.EXPECT().ListTeams(gomock.Any(), gomock.Eq(owner), gomock.Eq(repoName), gomock.Any()).Return([]*github.Team{
		{
			ID: p(teamId),
		},
	}, &github.Response{}, nil)
	sut := &pkg.GitHubTeamRepositoryFix{
		Owner:    owner,
		RepoName: repoName,
		TeamSlug: teamSlug,
	}
	err := sut.Apply()
	require.NoError(t, err)
}

func TestGithubRepositoryTeamFix_IntegrationTest(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("to run this test you must set env GITHUB_TOKEN first")
	}
	owner := readEssentialEnv(t, "GITHUB_TEAM_REPOSITORY_FIX_INTEGRATION_TEST_OWNER")
	repoName := readEssentialEnv(t, "GITHUB_TEAM_REPOSITORY_FIX_INTEGRATION_TEST_REPO_NAME")
	expectedTeam := readEssentialEnv(t, "GITHUB_TEAM_REPOSITORY_FIX_INTEGRATION_TEST_TEAM_SLUG")

	sut := &pkg.GitHubTeamRepositoryFix{
		Owner:    owner,
		RepoName: repoName,
		TeamSlug: expectedTeam,
	}
	err := sut.Apply()
	require.NoError(t, err)
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	teams, _, err := client.Repositories().ListTeams(context.Background(), owner, repoName, &github.ListOptions{
		PerPage: 100,
	})
	require.NoError(t, err)
	assert.True(t, linq.From(teams).AnyWith(func(i interface{}) bool {
		return *i.(*github.Team).Slug == expectedTeam
	}))
}
