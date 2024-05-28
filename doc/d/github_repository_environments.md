# `github_repository_environments` Data Block

The `github_repository_environments` data block in the `grept` tool is used to fetch the environments of a GitHub repository. This block can be used to retrieve the existing environments in a GitHub repository.

## Attributes

- `owner`: The owner of the GitHub repository.
- `repo_name`: The name of the GitHub repository.

## Exported Attributes

- `environments`: A list of environments that are present in the GitHub repository. Each environment is an object with the following attributes:
  - `name`: The name of the environment.
  - `node_id`: The node ID of the environment.

## Example

Here's an example of how to use the `github_repository_environments` data block in your configuration file:

```hcl
data "github_repository_environments" "example" {
  owner    = "owner_name"
  repo_name = "repo_name"
}