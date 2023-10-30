package pkg

import (
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"os"
)

func functions(baseDir string) map[string]function.Function {
	r := hcl2template.Functions(baseDir)
	r["env"] = function.New(&function.Spec{
		Description: "Read environment variable, return empty string if the variable is not set.",
		Params: []function.Parameter{
			function.Parameter{
				Name:         "key",
				Description:  "Environment variable name",
				Type:         cty.String,
				AllowUnknown: true,
				AllowMarked:  true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		RefineResult: func(builder *cty.RefinementBuilder) *cty.RefinementBuilder {
			return builder.NotNull()
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			key := args[0]
			if !key.IsKnown() {
				return cty.UnknownVal(cty.String), nil
			}
			return cty.StringVal(os.Getenv(key.AsString())), nil
		},
	})
	return r
}
