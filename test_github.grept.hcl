locals {
  repo_team      = "repo_team"
  expected_teams = [
    "testteam",
    local.repo_team
  ]
  expected_members  = ["lonegunmanb"]
  owner             = "grepttest"
  repo_name         = "demorepo"
  teams_need_create = {for team in data.github_team.expected_team : team.slug => team if team.team_name == ""}
  teams_need_attach = compliment(local.expected_teams, [
    for team in data.github_repository_teams.teams.teams :team.slug
  ])
}

data "github_team" expected_team {
  for_each = toset(local.expected_teams)
  owner    = "grepttest"
  slug     = each.value
}

data "github_repository_teams" teams {
  owner     = local.owner
  repo_name = local.repo_name
}

data "github_repository_collaborators" demo_repo {
  owner     = local.owner
  repo_name = local.repo_name
}

rule "must_be_true" all_teams_created {
  condition = alltrue([for team in data.github_team.expected_team : team.team_name != ""])
}

rule "must_be_true" repo_team_contain_maintainer {
  condition = toset(data.github_team.expected_team[local.repo_team].members) == toset(local.expected_members)
}

fix "github_team_members" repo_team_members {
  rule_ids  = [rule.must_be_true.repo_team_contain_maintainer.id]
  owner     = local.owner
  team_slug = local.repo_team
  dynamic "member" {
    for_each = local.expected_members
    content {
      username = member.value
      role     = "maintainer"
    }
  }
  depends_on = [fix.github_team.new_team]
}

rule "must_be_true" repo_should_not_contains_collaborator {
  condition = length(data.github_repository_collaborators.demo_repo.users) == 0
}

fix "github_team" new_team {
  for_each  = local.teams_need_create
  rule_ids  = [rule.must_be_true.all_teams_created.id]
  owner     = "grepttest"
  team_name = each.key
  privacy   = "closed"
}

rule "must_be_true" repo_has_essential_teams {
  condition = length(local.teams_need_attach) == 0
}

fix "github_team_repository" essential_teams {
  rule_ids  = [rule.must_be_true.repo_has_essential_teams.id]
  owner     = "grepttest"
  repo_name = local.repo_name
  dynamic "team" {
    for_each = local.teams_need_attach
    content {
      team_slug  = team.value
      permission = "admin"
    }
  }

  depends_on = [fix.github_team.new_team]
}