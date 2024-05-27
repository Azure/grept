# `github_repository_collaborators` Fix Block

The `github_repository_collaborators` fix block in the `grept` tool is used to manage the collaborators of a GitHub repository. This block can be used to ensure certain users are present or absent as collaborators in a GitHub repository.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `owner`: The owner of the GitHub repository.
- `repo_name`: The name of the GitHub repository.
- `collaborator`: A list of collaborators that must be present in the GitHub repository. Each collaborator is an object with the following attributes:
  - `user_name`: The username of the collaborator.
  - `permission`: The permission level of the collaborator. It can be one of the following: `pull`, `triage`, `push`, `maintain`, `admin`. The default value is `pull`.

## Exported Attributes

The `github_repository_collaborators` fix block does not export any attributes.

## Example

Here's an example of how to use the `github_repository_collaborators` fix block in your configuration file:

```hcl
fix "github_repository_collaborators" "example" {
  rule_ids = ["example_rule"]
  owner    = "owner_name"
  repo_name = "repo_name"
  collaborator {
    user_name = "collaborator1"
    permission = "push"
  }
  collaborator {
    user_name = "collaborator2"
    permission = "pull"
  }
}