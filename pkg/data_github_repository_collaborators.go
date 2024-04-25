package pkg

import (
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Data = &GitHubRepositoryTeamsDatasource{}

type User struct {
	Name    string `attribute:"name"`
	Company string `attribute:"company"`
	Id      int64  `attribute:"id"`
	Login   string `attribute:"login"`
	Email   string `attribute:"email"`
	Type    string `attribute:"type"`
	URL     string `attribute:"url"`
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
	opts := &github.ListCollaboratorsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		collaborators, resp, err := client.Repositories().ListCollaborators(g.Context(), g.Owner, g.RepoName, opts)
		if err != nil {
			return fmt.Errorf("cannot list collaborators for %s/%s: %s", g.Owner, g.RepoName, err.Error())
		}
		for _, c := range collaborators {
			g.Collaborators = append(g.Collaborators, User{
				Name:    value(c.Name),
				Company: value(c.Company),
				Id:      value(c.ID),
				Login:   value(c.Login),
				Email:   value(c.Email),
				Type:    value(c.Type),
				URL:     value(c.URL),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}

func value[T any](p *T) T {
	var d T
	if p == nil {
		return d
	}
	return *p
}
