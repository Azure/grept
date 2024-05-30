locals {
  ignored_items = toset([
    ".terraform.lock.hcl",
    ".terraformrc",
    "*_override.tf.json",
    "*_override.tf",
    "*.tfstate.*",
    "*.tfstate",
    "*.tfvars.json",
    "*.tfvars",
    "**/.terraform/*",
    "*tfplan*",
    "avm.tflint_example.hcl",
    "avm.tflint.hcl",
    "avm.tflint.merged.hcl",
    "avm.tflint_example.merged.hcl",
    "avmmakefile",
    "crash.*.log",
    "crash.log",
    "override.tf.json",
    "override.tf",
    "README-generated.md",
    "terraform.rc",
    ".DS_Store",
    "*.md.tmp",
  ])
}

data "git_ignore" "current_ignored_items" {}

rule "must_be_true" "essential_ignored_items" {
  condition = length(compliment(local.ignored_items, data.git_ignore.current_ignored_items.records)) == 0
}

fix "git_ignore" "ensure_ignore" {
  rule_ids = [rule.must_be_true.essential_ignored_items.id]
  exist   = local.ignored_items
}