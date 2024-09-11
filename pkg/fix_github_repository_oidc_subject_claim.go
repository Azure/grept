package pkg

import "github.com/Azure/golden"

var _ Fix = &GitHubRepositoryOidcSubjectClaimFix{}

type GitHubRepositoryOidcSubjectClaimFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner           string   `hcl:"owner"`
	RepoName        string   `hcl:"repo_name"`
	ClaimKeys       []string `hcl:"claim_keys"`
	ClaimUseDefault *bool    `hcl:"claim_use_default"`
}

func (g *GitHubRepositoryOidcSubjectClaimFix) Type() string {
	return "github_repository_oidc_subject_claim"
}

func (g *GitHubRepositoryOidcSubjectClaimFix) Apply() error {
	return nil
}
