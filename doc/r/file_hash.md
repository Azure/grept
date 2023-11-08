# `file_hash` Rule Block

The `file_hash` rule block in the Grept tool is used to enforce that a certain file in the repository has a specific hash. This can be used to ensure the integrity of important files.

## Attributes

- `path`: The path of the file to check.
- `hash`: The expected hash of the file.
- `error_message`: The error message that will be displayed if the rule fails.

## Exported Attributes

- `id`: The ID of the rule. This is automatically generated and should not be set by the user.

## Example

Here's an example of how to use the `file_hash` rule block in your configuration file:

```hcl
rule "file_hash" "example" {
  path          = "/path/to/file"
  hash          = "ab3c1234f5678901e2d34fa5678b90cd"
  error_message = "The file /path/to/file must have the correct hash"
}
```

This will enforce that the file at `/path/to/file` has the hash `ab3c1234f5678901e2d34fa5678b90cd`. If it doesn't, the rule will fail and the error message "The file /path/to/file must have the correct hash" will be displayed. The ID of the rule will be automatically generated.