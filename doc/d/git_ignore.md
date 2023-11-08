# `git_ignore` Data Block

The `git_ignore` data block in the grept tool is used to load data from a `.gitignore` file. It reads the file and stores the ignore patterns as records.

## Attributes

The `git_ignore` data block does not support any attributes. It automatically loads the `.gitignore` file from the root of the repository.

## Exported Attributes

- `records`: This attribute is a list of strings, where each string is a line from the `.gitignore` file. This excludes any lines that are comments (i.e., start with `#`) or empty lines.

## Example

Here's an example of how to use the `git_ignore` data block in your configuration file:

```hcl
data "git_ignore" "example" {
  # No attributes are needed
}
```

You can then access the records exported by this block in your rules or fixes. For example:

```hcl
rule "must_be_true" "example" {
  condition = contains(data.git_ignore.example.records, "/bin")
  error_message = "The /bin directory should be ignored"
}
```

This will check if the `/bin` directory is listed in the `.gitignore` file, and return an error if it's not.