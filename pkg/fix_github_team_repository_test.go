package pkg_test

import (
	"context"
	"fmt"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/require"
)

var _ githubclient.TeamsClient = &fakeRepoTeams{}
var _ githubclient.RepositoriesClient = &fakeRepoTeams{}
var _ githubclient.OrganizationsClient = &fakeRepoTeams{}

type fakeRepoTeams struct {
	teams     map[string]*github.Team
	repoTeams map[string]int64
	orgs      map[string]*github.Organization
}

func (m *fakeRepoTeams) Get(ctx context.Context, org string) (*github.Organization, *github.Response, error) {
	return m.orgs[org], &github.Response{}, nil
}

func (m *fakeRepoTeams) ListTeams(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.Team, *github.Response, error) {
	teamId, ok := m.repoTeams[fmt.Sprintf("%s/%s", owner, repo)]
	if ok {
		return []*github.Team{
			&github.Team{ID: p(teamId)},
		}, &github.Response{}, nil
	}
	return []*github.Team{}, &github.Response{}, nil
}

func (m *fakeRepoTeams) ListCollaborators(ctx context.Context, owner, repo string, opts *github.ListCollaboratorsOptions) ([]*github.User, *github.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *fakeRepoTeams) GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error) {
	return m.teams[fmt.Sprintf("%s/%s", org, slug)], &github.Response{}, nil
}

func (m *fakeRepoTeams) AddTeamRepoByID(ctx context.Context, orgID, teamID int64, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error) {
	m.repoTeams[fmt.Sprintf("%s/%s", owner, repo)] = teamID
	return &github.Response{}, nil
}

func (m *fakeRepoTeams) RemoveTeamRepoByID(ctx context.Context, orgID, teamID int64, owner, repo string) (*github.Response, error) {
	delete(m.repoTeams, fmt.Sprintf("%s/%s", owner, repo))
	return &github.Response{}, nil
}

func TestGithubRepositoryTeamFix(t *testing.T) {
	owner := "Azure"
	repoName := "testrepo"
	expectedTeam := "testteam"
	teamId := int64(123)
	mock := &fakeRepoTeams{
		orgs: map[string]*github.Organization{
			owner: &github.Organization{
				ID:   p(int64(456)),
				Name: p(owner),
			},
		},
		teams: map[string]*github.Team{
			fmt.Sprintf("%s/%s", owner, expectedTeam): &github.Team{
				ID:   p(teamId),
				Name: p(expectedTeam),
			},
		},
		repoTeams: map[string]int64{},
	}
	stub := gostub.Stub(&githubclient.GetGithubClient, func() (*githubclient.Client, error) {
		return &githubclient.Client{
			Repositories: func() githubclient.RepositoriesClient {
				return mock
			},
			Organizations: func() githubclient.OrganizationsClient {
				return mock
			},
			Teams: func() githubclient.TeamsClient {
				return mock
			},
		}, nil
	})
	defer stub.Reset()
	sut := &pkg.GitHubTeamRepositoryFix{
		Owner:    owner,
		RepoName: repoName,
		Teams: []pkg.TeamRepositoryBinding{
			{
				TeamSlug:   expectedTeam,
				Permission: "pull",
			},
		},
	}
	err := sut.Apply()
	require.NoError(t, err)
	assert.Len(t, mock.repoTeams, 1)
	actualTeamId := mock.repoTeams[fmt.Sprintf("%s/%s", owner, repoName)]
	assert.Equal(t, teamId, actualTeamId)
}
