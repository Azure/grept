//go:generate mockgen -destination ../mock_repositories_client_test.go -package pkg_test . RepositoriesClient,TeamsClient,OrganizationsClient
package githubclient

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v61/github"
	"net/http"
	"os"
	"strconv"
)

type Client struct {
	*github.Client
	Repositories  func() RepositoriesClient
	Teams         func() TeamsClient
	Organizations func() OrganizationsClient
}

var GetGithubClient = func() (*Client, error) {
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		return newClient(github.NewClient(nil).WithAuthToken(githubToken)), nil
	}
	ghAppIntegrationIdRaw := os.Getenv("GITHUB_APP_INTEGRATION_ID")
	integrationId, err := strconv.ParseInt(ghAppIntegrationIdRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("incorrect GITHUB_APP_INTEGRATION_ID: %s", ghAppIntegrationIdRaw)
	}
	installationIdRaw := os.Getenv("GITHUB_APP_INSTALLATION_ID")
	installationId, err := strconv.ParseInt(installationIdRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("incorrect GITHUB_APP_INSTALLATION_ID: %s", installationIdRaw)
	}
	privateKey := os.Getenv("GITHUB_APP_PRIVATEKEY")
	if privateKey == "" {
		return nil, fmt.Errorf("must set env GITHUB_APP_PRIVATEKEY")
	}
	itr, err := ghinstallation.New(http.DefaultTransport, integrationId, installationId, []byte(privateKey))
	if err != nil {
		return nil, err
	}
	return newClient(github.NewClient(&http.Client{Transport: itr})), nil
}

func newClient(c *github.Client) *Client {
	r := &Client{
		Client: c,
	}
	r.Organizations = func() OrganizationsClient {
		return c.Organizations
	}
	r.Repositories = func() RepositoriesClient {
		return c.Repositories
	}
	r.Teams = func() TeamsClient {
		return c.Teams
	}

	return r
}

type RepositoriesClient interface {
	ListTeams(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.Team, *github.Response, error)
	ListCollaborators(ctx context.Context, owner, repo string, opts *github.ListCollaboratorsOptions) ([]*github.User, *github.Response, error)
}

type TeamsClient interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error)
	AddTeamRepoByID(ctx context.Context, orgID, teamID int64, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error)
	RemoveTeamRepoByID(ctx context.Context, orgID, teamID int64, owner, repo string) (*github.Response, error)
}

type OrganizationsClient interface {
	Get(ctx context.Context, org string) (*github.Organization, *github.Response, error)
}
