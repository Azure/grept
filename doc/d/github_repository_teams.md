# `github_repository_teams` Data Block

The `github_repository_teams` data block in the `grept` tool is used to fetch the teams of a GitHub repository. This block can be used to retrieve the existing teams in a GitHub repository.

## Attributes

- `owner`: The owner of the GitHub repository.
- `repo_name`: The name of the GitHub repository.

## Exported Attributes

- `teams`: A list of teams that are present in the GitHub repository. Each team is an object with the following attributes:
  - `name`: The name of the team.
  - `slug`: The slug of the team.
  - `permission`: The permission of the team.

## Example

Here's an example of how to use the `github_repository_teams` data block in your configuration file:

```hcl
data "github_repository_teams" "example" {
  owner    = "owner_name"
  repo_name = "repo_name"
}