# `must_be_true` Rule Block

The `must_be_true` rule block in the `grept` tool is used to enforce that a certain expression evaluates to true. This can be used to set up various conditions that must be met in the repository.

## Attributes

- `condition`: The expression that should evaluate to true.
- `error_message`: The error message that will be displayed if the rule fails.

## Exported Attributes

- `id`: The ID of the rule. This is automatically generated and should not be set by the user.

## Example

Here's an example of how to use the `must_be_true` rule block in your configuration file:

```hcl
rule "must_be_true" "example" {
  condition     = (2 + 2) == 4
  error_message = "The expression must be true"
}
```

This will enforce that the expression `(2 + 2) == 4` evaluates to true. If it doesn't, the rule will fail and the error message "The expression must be true" will be displayed. The ID of the rule will be automatically generated.
