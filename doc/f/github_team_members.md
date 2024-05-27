# `github_team_members` Fix Block

The `github_team_members` fix block in the `grept` tool is used to manage the members of a team in a GitHub organization. This block can be used to ensure certain users are present with specific roles in a team.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `owner`: The owner of the GitHub organization.
- `team_slug`: The slug of the team.
- `member`: A list of members that must be present in the team. Each member is an object with the following attributes:
  - `username`: The username of the member.
  - `role`: (optional) The role of the member. It can be one of the following: `member`, `maintainer`. The default value is `member`.

## Exported Attributes

The `github_team_members` fix block does not export any attributes.

## Example

Here's an example of how to use the `github_team_members` fix block in your configuration file:

```hcl
fix "github_team_members" "example" {
  rule_ids   = ["example_rule"]
  owner      = "owner_name"
  team_slug  = "team_slug"
  member {
    username = "member1"
    role     = "maintainer"
  }
  member {
    username = "member2"
    role     = "member"
  }
}