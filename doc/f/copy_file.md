# `copy_file` Fix Block

The `copy_file` fix block in the `grept` tool is used to copy a file from a source path to a destination path. This can be used to enforce certain file presence rules or to ensure a specific version of a file is present.

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `src`: The source file path that should be copied.
- `dest`: The destination file path where the source file should be copied to.

## Exported Attributes

The `copy_file` fix block does not export any attributes.

## Example

Here's an example of how to use the `copy_file` fix block in your configuration file:

```hcl
fix "copy_file" "example" {
  rule_ids = ["example_rule"]
  src      = "/path/to/source/file"
  dest     = "/path/to/destination/file"
}
```

This will copy the file at `/path/to/source/file` to `/path/to/destination/file` if the rule with ID `example_rule` fails.