# `github_repository_collaborators` Data Block

The `github_repository_collaborators` data block in the `grept` tool is used to fetch the collaborators of a GitHub repository. This block can be used to retrieve the existing collaborators in a GitHub repository.

## Attributes

- `owner`: The owner of the GitHub repository.
- `repo_name`: The name of the GitHub repository.

## Exported Attributes

- `collaborators`: A list of collaborators that are present in the GitHub repository. Each collaborator is an object with the following attributes:
  - `name`: The name of the collaborator.
  - `company`: The company of the collaborator.
  - `id`: The ID of the collaborator.
  - `login`: The login of the collaborator.
  - `email`: The email of the collaborator.
  - `type`: The type of the collaborator.
  - `url`: The URL of the collaborator.
  - `permissions`: The permissions of the collaborator.

## Example

Here's an example of how to use the `github_repository_collaborators` data block in your configuration file:

```hcl
data "github_repository_collaborators" "example" {
  owner    = "owner_name"
  repo_name = "repo_name"
}