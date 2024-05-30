# This policy is designed to manage collaborators for a specific GitHub repository.
# It uses the `github_repository_collaborators` data source to fetch the current collaborators of the repository.
# The `rule` block checks if a specific collaborator (defined by `collaborator_name` variable) is present in the repository's collaborators list.
# The `fix` block is used to add the collaborator to the repository with 'push' permission if they are not already a collaborator.
# The `owner` and `repo_name` variables define the repository for which the policy is applied.
# The `collaborator_name` variable defines the collaborator to be checked/added.
variable "owner" {
  type    = string
  default = "Azure"
}

variable "repo_name" {
  type    = string
  default = "terraform-azurerm-aks"
}

variable "collaborator_name" {
  type = string
}

data "github_repository_collaborators" "current_collaborators" {
  owner     = var.owner
  repo_name = var.repo_name
}

rule "must_be_true" "collaborator_exists" {
  condition = toset([
    for user in data.github_repository_collaborators.current_collaborators.users :user.login
  ]) == toset([var.collaborator_name])
  error_message = "Collaborators are not as expected"
}

fix "github_repository_collaborators" "add_collaborator" {
  rule_ids  = [rule.must_be_true.collaborator_exists.id]
  owner     = var.owner
  repo_name = var.repo_name
  collaborator {
    user_name  = var.collaborator_name
    permission = "push"
  }
}