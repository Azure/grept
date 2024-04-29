package pkg_test

import (
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGithubTeamDatasource(t *testing.T) {
	owner := "Azure"
	expectedTeam := "testteam"
	teamId := int64(123)
	mock := &fakeRepoTeams{
		teams: map[string]*github.Team{
			fmt.Sprintf("%s/%s", owner, expectedTeam): &github.Team{
				ID:          p(teamId),
				Name:        p(expectedTeam),
				Description: p("description"),
				Permission:  p("permission"),
				NodeID:      p("node_id"),
				Privacy:     p("privacy"),
				Slug:        p(expectedTeam),
			},
		},
		repoTeams: map[string]int64{},
	}
	stub := gostub.Stub(&githubclient.GetGithubClient, func() (*githubclient.Client, error) {
		return &githubclient.Client{
			Teams: func() githubclient.TeamsClient {
				return mock
			},
		}, nil
	})
	defer stub.Reset()
	sut := &pkg.GitHubTeamDatasource{
		Owner: owner,
		Slug:  expectedTeam,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.Equal(t, teamId, sut.TeamId)
	assert.Equal(t, expectedTeam, sut.TeamName)
	assert.Equal(t, "privacy", sut.Privacy)
	assert.Equal(t, "node_id", sut.NodeId)
	assert.Equal(t, expectedTeam, sut.Slug)
	assert.Equal(t, "permission", sut.Permission)
	assert.Equal(t, owner, sut.Owner)
}

func TestGithubTeamRead_IntegrationTest(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("to run this test you must set env GITHUB_TOKEN first")
	}
	owner := readEssentialEnv(t, "GITHUB_TEAM_DATASOURCE_INTEGRATION_TEST_OWNER")
	expectedTeam := readEssentialEnv(t, "GITHUB_TEAM_DATASOURCE_INTEGRATION_TEST_EXPECTED_TEAM")

	sut := &pkg.GitHubTeamDatasource{
		Owner: owner,
		Slug:  expectedTeam,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.Equal(t, expectedTeam, sut.TeamName)
}
