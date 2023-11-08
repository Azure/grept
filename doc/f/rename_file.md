# `rename_file` Fix Block

The `rename_file` fix block in the `grept` tool is used to rename a local file. This can be used to enforce certain file naming rules.

## Attributes

- `rule_id`: The ID of the rule this fix is associated with.
- `old_name`: The current name of the file to be renamed.
- `new_name`: The new name for the file.

## Exported Attributes

The `rename_file` fix block does not export any attributes.

## Example

Here's an example of how to use the `rename_file` fix block in your configuration file:

```hcl
fix "rename_file" "example" {
  rule_id  = "example_rule"
  old_name = "/path/to/old_file"
  new_name = "/path/to/new_file"
}
```

This will rename the file at `/path/to/old_file` to `/path/to/new_file` if the rule with ID `example_rule` fails.