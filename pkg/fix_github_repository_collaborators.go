package pkg

import (
	"context"
	"fmt"
	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
	"net/http"
	"strings"
)

var _ Fix = &GitHubRepositoryCollaboratorsFix{}

type CollaboratorForRepositoryCollaboratorsFix struct {
	Name       string `hcl:"user_name"`
	Permission string `hcl:"permission" default:"pull" validate:"oneof=pull triage push maintain admin"`
}

type GitHubRepositoryCollaboratorsFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner         string                                      `hcl:"owner"`
	RepoName      string                                      `hcl:"repo_name"`
	Collaborators []CollaboratorForRepositoryCollaboratorsFix `hcl:"collaborator,block"`
}

func (g *GitHubRepositoryCollaboratorsFix) Type() string {
	return "github_repository_collaborators"
}

func (g *GitHubRepositoryCollaboratorsFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}

	collaborators, err := listRepositoryCollaborators(client, g.Context(), g.Owner, g.RepoName)
	if err != nil {
		return err
	}
	for _, collaborator := range collaborators {
		// Delete any pending invitations
		invitation, err := FindRepoInvitation(client, g.Context(), g.Owner, g.RepoName, collaborator.Login)
		if err != nil {
			if ghErr, ok := err.(*github.ErrorResponse); !ok || ghErr.Response.StatusCode != http.StatusNotFound {
				return fmt.Errorf("error reading repo invitation for %s/%s: %+v", g.Owner, g.RepoName, err)
			}
		}
		if invitation != nil {
			_, err = client.Repositories.DeleteInvitation(g.Context(), g.Owner, g.RepoName, invitation.GetID())
			return err
		}
		_, err = client.Repositories.RemoveCollaborator(g.Context(), g.Owner, g.RepoName, collaborator.Login)
		if err != nil {
			return fmt.Errorf("cannot remove collaborator %s from %s/%s", collaborator.Login, g.Owner, g.RepoName)
		}
	}
	for _, member := range g.Collaborators {
		_, _, err = client.Repositories.AddCollaborator(g.Context(),
			g.Owner, g.RepoName, member.Name,
			&github.RepositoryAddCollaboratorOptions{
				Permission: member.Permission,
			})
		if err != nil {
			return fmt.Errorf("cannot add collaborator %s to %s/%s", member.Name, g.Owner, g.RepoName)
		}
	}
	return nil
}

func FindRepoInvitation(client *githubclient.Client, ctx context.Context, owner, repo, collaborator string) (*github.RepositoryInvitation, error) {
	opt := &github.ListOptions{PerPage: 100}
	for {
		invitations, resp, err := client.Repositories.ListInvitations(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}

		for _, i := range invitations {
			if strings.EqualFold(i.GetInvitee().GetLogin(), collaborator) {
				return i, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil, nil
}
