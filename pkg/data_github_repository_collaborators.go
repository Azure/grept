package pkg

import (
	"context"
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Data = &GitHubRepositoryTeamsDatasource{}

type User struct {
	Name        string   `attribute:"name"`
	Company     string   `attribute:"company"`
	Id          int64    `attribute:"id"`
	Login       string   `attribute:"login"`
	Email       string   `attribute:"email"`
	Type        string   `attribute:"type"`
	URL         string   `attribute:"url"`
	Permissions []string `attribute:"permissions"`
}

type GitHubRepositoryCollaboratorsDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner         string `hcl:"owner" json:"owner"`
	RepoName      string `hcl:"repo_name" json:"repo_name"`
	Collaborators []User `attribute:"users"`
}

func (g *GitHubRepositoryCollaboratorsDatasource) Type() string {
	return "github_repository_collaborators"
}

func (g *GitHubRepositoryCollaboratorsDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	collaborators, err := listRepositoryCollaborators(client, g.Context(), g.Owner, g.RepoName)
	if err != nil {
		return err
	}
	g.Collaborators = collaborators
	return nil
}

func listRepositoryCollaborators(client *githubclient.Client, ctx context.Context, owner, repoName string) ([]User, error) {
	var r []User
	opts := &github.ListCollaboratorsOptions{
		Affiliation: "direct",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		collaborators, resp, err := client.Repositories.ListCollaborators(ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("cannot list collaborators for %s/%s: %s", owner, repoName, err.Error())
		}
		for _, c := range collaborators {
			user := User{
				Name:    value(c.Name),
				Company: value(c.Company),
				Id:      value(c.ID),
				Login:   value(c.Login),
				Email:   value(c.Email),
				Type:    value(c.Type),
				URL:     value(c.URL),
			}
			for p, enabled := range c.Permissions {
				if enabled {
					user.Permissions = append(user.Permissions, p)
				}
			}
			r = append(r, user)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return r, nil
}

func value[T any](p *T) T {
	var d T
	if p == nil {
		return d
	}
	return *p
}
