# This example demonstrate how to ensure that a set of files are in sync with their remote versions.
# If you've set env `GITHUB_TOKEN` to increase GitHub API rate limitation, you must grant this token SSO permission on Azure org.
locals {
  synced_files = toset([
    "CODE_OF_CONDUCT.md",
    "LICENSE",
    "SECURITY.md",
  ])
  url_prefix        = "https://raw.githubusercontent.com/Azure/terraform-azurerm-avm-template/main"
  common_http_headers = local.github_token=="" ? {} : {
    Authorization = "Bearer ${local.github_token}"
  }
  github_token = env("GITHUB_TOKEN")
}

data "http" "synced_files" {
  for_each = local.synced_files

  request_headers = merge({}, local.common_http_headers)
  url             = "${local.url_prefix}/${each.value}"
}

rule "file_hash" "synced_files" {
  for_each = local.synced_files

  glob = each.value
  hash = sha1(data.http.synced_files[each.value].response_body)
}

fix "local_file" "synced_files" {
  for_each = local.synced_files

  rule_ids = [rule.file_hash.synced_files[each.value].id]
  paths    = [each.value]
  content  = data.http.synced_files[each.value].response_body
}