# `github_team_repository` Fix Block

The `github_team_repository` fix block in the `grept` tool is used to manage the repository permissions of a team in a GitHub organization. This block can be used to ensure certain teams have specific permissions on a repository.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `owner`: The owner of the GitHub organization.
- `repo_name`: The name of the GitHub repository.
- `team`: A list of teams that must have permissions on the repository. Each team is an object with the following attributes:
  - `team_slug`: The slug of the team.
  - `permission`: (optional) The permission level of the team on the repository. It can be one of the following: `pull`, `triage`, `push`, `maintain`, `admin`. The default value is `pull`.

## Exported Attributes

The `github_team_repository` fix block does not export any attributes.

## Example

Here's an example of how to use the `github_team_repository` fix block in your configuration file:

```hcl
fix "github_team_repository" "example" {
  rule_ids   = ["example_rule"]
  owner      = "owner_name"
  repo_name  = "repo_name"
  team {
    team_slug = "team_slug1"
    permission = "push"
  }
  team {
    team_slug = "team_slug2"
    permission = "pull"
  }
}