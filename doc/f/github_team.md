# `github_team` Fix Block

The `github_team` fix block in the `grept` tool is used to manage the teams of a GitHub organization. This block can be used to ensure a certain team is present with specific configurations in a GitHub organization.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `owner`: The owner of the GitHub organization.
- `team_name`: The name of the team.
- `description`: (optional) The description of the team.
- `privacy`: (optional) The privacy level of the team. It can be one of the following: `secret`, `closed`. The default value is `secret`.
- `parent_team_id`: (optional) The ID of the parent team. The default value is `-1`, which means no parent team.
- `ldap_dn`: (optional) The LDAP distinguished name of the team.
- `create_default_maintainer`: (optional) Whether to create a default maintainer for the team. The default value is `false`.

## Exported Attributes

The `github_team` fix block does not export any attributes.

## Example

Here's an example of how to use the `github_team` fix block in your configuration file:

```hcl
fix "github_team" "example" {
  rule_ids   = ["example_rule"]
  owner      = "owner_name"
  team_name  = "team_name"
  description = "team_description"
  privacy    = "closed"
  parent_team_id = 123456
  ldap_dn = "ldap_dn"
  create_default_maintainer = false
}