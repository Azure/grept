# `rm_local_file` Fix Block

The `rm_local_file` fix block in the `grept` tool is used to remove a local file. This can be used to enforce rules about which files should not exist in the repository.

## Attributes

- `rule_id`: The ID of the rule this fix is associated with.
- `path`: The path of the file to be removed.

## Exported Attributes

The `rm_local_file` fix block does not export any attributes.

## Example

Here's an example of how to use the `rm_local_file` fix block in your configuration file:

```hcl
fix "rm_local_file" "example" {
  rule_id = "example_rule"
  path    = "/path/to/file"
}
```

This will remove the file at `/path/to/file` if the rule with ID `example_rule` fails.