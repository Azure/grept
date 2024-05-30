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