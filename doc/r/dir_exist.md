# `dir_exist` Rule Block

The `dir_exist` rule block in the `grept` tool is used to enforce that a certain directory exists within the repository.

## Attributes

- `dir`: The path of the directory that should exist.
- `fail_on_exist`: Set this attribute to `true` will fail the check once a directory is found. Defaults to `false`.

## Exported Attributes

- `id`: The ID of the rule.

## Example

Here's an example of how to use the `dir_exist` rule block in your configuration file:

```hcl
rule "dir_exist" "example" {
  dir = "/path/to/dir"
}
```

This will enforce that the directory at `/path/to/dir` exists in the repository. If it doesn't, the rule will fail and the error message "The directory /path/to/dir must exist" will be displayed. The ID of the rule will be automatically generated.

```hcl
rule "dir_exist" "example" {
  dir           = "/path/to/dir"
  fail_on_exist = true
}
```

This will enforce that the directory at `/path/to/dir` not exist in the repository. If it does, the rule will fail.