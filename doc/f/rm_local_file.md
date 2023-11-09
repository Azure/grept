# `rm_local_file` Fix Block

The `rm_local_file` fix block in the `grept` tool is used to remove a local file. This can be used to enforce rules about which files should not exist in the repository.

## Attributes

- `rule_id`: The ID of the rule this fix is associated with.
- `paths`: The list of paths of the files or directories to be removed. If a path points a directory, all sub folders and files in that directory would be deleted. If a path points to a non-exist file or directory, no error would be thrown.

## Exported Attributes

The `rm_local_file` fix block does not export any attributes.

## Example

Here's an example of how to use the `rm_local_file` fix block in your configuration file:

```hcl
fix "rm_local_file" "example" {
  rule_id  = "example_rule"
  paths    = ["/path/to/file"]
}
```

This will remove the file at `/path/to/file` if the rule with ID `example_rule` fails.