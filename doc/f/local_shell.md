# `local_shell` Fix Block

The `local_shell` fix block in the `grept` tool is used to execute a shell command or script as a fix. This can be used to perform various actions such as modifying files, changing permissions, or any other operation that can be performed via a shell command. Any non-zero return value of the script would cause error.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `execute_command`: The command used to execute the script. Defaults to `['/bin/sh', '-c']`.
- `inline_shebang`: The shebang line to be used when executing inline scripts. Defaults to `/bin/sh -e`. Must be set along with `inlines`.
- `inlines`: A list of inline scripts to be executed. Must not be set along with `script` or `remote_script`.
- `script`: Path to a local script file to be executed. Must not be set along with `inlines` or `remote_script`.
- `remote_script`: URL of a remote script to be downloaded and executed. Must not be set along with `inlines` or `script`.
- `only_on`: A list of operating systems where the fix should be applied. Valid values are `windows`, `linux`, `darwin`, `openbsd`, `netbsd`, `freebsd`, `dragonfly`, `android`, `solaris`, `plan9`. If the current os doesn't in this list, `local_shell` fix would return directly without error.
- `env`: A map of environment variables to be set when executing the script.

## Exported Attributes

The `local_shell` fix block does not export any attributes.

## Example

Here's an example of how to use the `local_shell` fix block in your configuration file:

```hcl
fix "local_shell" "example" {
  rule_ids = ["example_rule"]
  inlines  = ["echo 'This is an inline script'", "ls -l"]
}
```

This will execute the inline scripts `echo 'This is an inline script'` and `ls -l` if the rule with ID `example_rule` fails.

Alternatively, you can provide a local or remote script to be executed:

```hcl
fix "local_shell" "example" {
  rule_ids = ["example_rule"]
  script   = "/path/to/local/script.sh"
}

fix "local_shell" "example" {
  rule_ids       = ["example_rule"]
  remote_script  = "http://example.com/script.sh"
}
```

These will execute the local script at `/path/to/local/script.sh` or the remote script at `http://example.com/script.sh` respectively if the rule with ID `example_rule` fails.

You can also set environment variables for the script:

```hcl
fix "local_shell" "example" {
  rule_ids = ["example_rule"]
  script   = "/path/to/local/script.sh"
  env = {
    "VAR1" = "value1"
    "VAR2" = "value2"
  }
}
```

This will set the environment variables `VAR1` and `VAR2` before executing the script at `/path/to/local/script.sh` if the rule with ID `example_rule` fails."
