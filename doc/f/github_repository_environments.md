# `github_repository_environments` Fix Block

The `github_repository_environments` fix block in the `grept` tool is used to manage the [environments of a GitHub repository](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment). This block can be used to ensure certain environments are present with specific configurations in a GitHub repository.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `owner`: The owner of the GitHub repository.
- `repo_name`: The name of the GitHub repository.
- `environment`: A list of environments that must be present in the GitHub repository. Each environment is an object with the following attributes:
  - `name`: The name of the environment.
  - `can_admins_bypass`: Whether admins can bypass the required reviewers. Defaults to `true`.
  - `prevent_self_review`: Whether to prevent the creator of a deployment from approving their own deployment. Defaults to `false`.
  - `wait_timer`: (optional) The amount of time (in minutes) to wait before auto-merging a deployment. Must be between 0 and 43200.
  - `reviewer`: A list of reviewers for the environment. Each reviewer is an object with the following attributes:
    - `team_id`: (optional) The ID of the team reviewer.
    - `user_id`: (optional) The ID of the user reviewer.
  - `deployment_branch_policy`: The branch policy for deployments. It is an object with the following attributes:
    - `protected_branches`: Whether only branches with branch protection rules can deploy to this environment. If `protected_branches` is `true`, `custom_branch_policies` must be `false`; if `protected_branches` is `false`, `custom_branch_policies` must be `true`.
    - `custom_branch_policies`: Whether only branches that match the specified name patterns can deploy to this environment. If `custom_branch_policies` is `true`, `protected_branches` must be `false`; if `custom_branch_policies` is `false`, `protected_branches` must be `true`.

## Exported Attributes

The `github_repository_environments` fix block does not export any attributes.

## Example

Here's an example of how to use the `github_repository_environments` fix block in your configuration file:

```hcl
fix "github_repository_environments" "example" {
  rule_ids = ["example_rule"]
  owner    = "owner_name"
  repo_name = "repo_name"
  environment {
    name = "environment1"
    can_admins_bypass = true
    prevent_self_review = false
    wait_timer = 10
    reviewer {
      team_id = 123456
    }
    reviewer {
      user_id = 654321
    }
    deployment_branch_policy {
      protected_branches = true
      custom_branch_policies = false
    }
  }
}