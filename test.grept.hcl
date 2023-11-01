rule file_hash license {
  glob = "LICENSE"
  hash = sha1(data.http.mit_license.response_body)
}

rule must_be_true test {
  condition = env("OS") == "windows"
}
#
#fix local_file warning {
#  rule_id = rule.must_be_true.test.id
#  paths = ["warning"]
#  content = "not windows"
#}

fix local_file license {
  rule_id = rule.file_hash.license.id
  paths = ["LICENSE"]
  content = data.http.mit_license.response_body
}