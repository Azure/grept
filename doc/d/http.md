# `http` Data Block

The `http` data block in the `grept` tool is used to make HTTP requests and load the response data. You can specify the URL, request method, request body, and headers.

## Attributes

- `url`: The URL to send the HTTP request to.
- `method`: The HTTP method to use for the request (`GET`, `POST`, etc.). Default is `GET`.
- `request_body`: The request body to send with the request, if applicable.
- `request_headers`: A map of request headers to include with the request.
- `retry_max`: The maximum number of times to retry the request in case of failure. Default is `4`.

## Exported Attributes

- `response_body`: The body of the HTTP response.
- `response_headers`: A map of the headers in the HTTP response.
- `status_code`: The status code of the HTTP response.

## Example

Here's an example of how to use the `http` data block in your configuration file:

```hcl
data "http" "example" {
  url          = "https://example.com/api"
  method       = "POST"
  request_body = jsonencode({
    query = "example"
  })
  request_headers = {
    "Content-Type" = "application/json"
  }
}
```

You can then access the response data exported by this block in your rules or fixes. For example:

```hcl
rule "must_be_true" "example" {
  condition     = data.http.example.status_code == 200
  error_message = "API request must return status code 200"
}
```

This will check if the HTTP request returned a status code of 200, and return an error if it didn't.