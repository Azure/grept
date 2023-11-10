# `yaml_transform` Fix Block

The `yaml_transform` fix block in the `grept` tool is used to manipulate the contents of a YAML file as a fix. It allows specifying a list of transformations to be applied to the file.

**It only supports set a string attribute for now.**

## Attributes

- `rule_ids`: The ID list of the rules this fix is associated with. Any rule check failure would trigger this fix.
- `file_path`: The path to the YAML file to be transformed. The file should have a `.yaml` or `.yml` extension.
- `transform`: A list of transformations to be applied to the YAML file. Each transformation is a block with the following attributes:
  - `yaml_path`: The path to the node in the YAML file to be transformed. The path uses JSON Pointer syntax.
  - `string_value`: The new value to be set for the node.

## Exported Attributes

The `yaml_transform` fix block does not export any attributes.

## Example

Here's an example of how to use the `yaml_transform` fix block in your configuration file:

```hcl
fix "yaml_transform" "example" {
  rule_ids   = ["example_rule"]
  file_path  = "./path/to/file.yaml"
  transform {
    yaml_path    = "/path/to/node"
    string_value = "new value"
  }
  transform {
    yaml_path    = "/path/to/another/node"
    string_value = "another new value"
  }
}
```

This will change the value of the node at `/path/to/node` to `"new value"` and the node at `/path/to/another/node` to `"another new value"` in the YAML file at `./path/to/file.yaml` if the rule with ID `example_rule` fails.

You can also use `~{}` syntax to select elements from a collection to transform, for yaml:

```yaml
name: pr-check
on:
  workflow_dispatch:
  pull_request:
    types: [ 'opened', 'synchronize' ]
  push:  
    branches:  
      - main

permissions:
  contents: write
  pull-requests: read
  statuses: write
  security-events: write
  actions: read

jobs:
  prepr-check:
    runs-on: ubuntu-latest
    steps:
      - name: checkout
        uses: actions/checkout@f43a0e5ff2bd294095638e18286ca9a3d1956744 #v3.6.0
      - name: pr-check
        run: |
          docker run --rm -v $(pwd):/src -w /src -e SKIP_CHECKOV -e GITHUB_TOKEN mcr.microsoft.com/azterraform:latest make pr-check
      - name: Set up Go
        uses: actions/setup-go@6edd4406fa81c3da01a34fa6f6343087c207a568 #3.5.0
        with:
          go-version: 1.21.3
```

You can set `checkout` step's action version like this:

```hcl
fix "yaml_transform" "example" {
  rule_ids  = ["example_rule"]
  file_path = "./path/to/file.yaml"
  transform {
    yaml_path    = "/jobs/prepr-check/steps/~{"name":"checkout"}/uses"
    string_value = "actions/checkout@v3.7.0"
  }
}
```

The transformed yaml file would be:

```yaml
name: pr-check
on:
  workflow_dispatch:
  pull_request:
    types: [ 'opened', 'synchronize' ]
  push:  
    branches:  
      - main

permissions:
  contents: write
  pull-requests: read
  statuses: write
  security-events: write
  actions: read

jobs:
  prepr-check:
    runs-on: ubuntu-latest
    steps:
      - name: checkout
        uses: actions/checkout@v3.7.0
      - name: pr-check
        run: |
          docker run --rm -v $(pwd):/src -w /src -e SKIP_CHECKOV -e GITHUB_TOKEN mcr.microsoft.com/azterraform:latest make pr-check
      - name: Set up Go
        uses: actions/setup-go@6edd4406fa81c3da01a34fa6f6343087c207a568 #3.5.0
        with:
          go-version: 1.21.3
```

Please check out [VMWare yaml-jsonpointer](https://github.com/vmware-archive/yaml-jsonpointer) for more details.
