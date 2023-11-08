# `dir_exist` Rule Block

The `dir_exist` rule block in the `grept` tool is used to enforce that a certain directory exists within the repository.

## Attributes

- `path`: The path of the directory that should exist.
- `error_message`: The error message that will be displayed if the rule fails.

## Exported Attributes

- `id`: The ID of the rule.

## Example

Here's an example of how to use the `dir_exist` rule block in your configuration file:

```hcl
rule "dir_exist" "example" {
  id             = "example_rule"
  path           = "/path/to/dir"
  error_message  = "The directory /path/to/dir must exist"
}
```

This will enforce that the directory at `/path/to/dir` exists in the repository. If it doesn't, the rule will fail and the error message "The directory /path/to/dir must exist" will be displayed. The ID of the rule will be automatically generated.