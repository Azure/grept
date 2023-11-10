package pkg

import (
	"fmt"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/timandy/routine"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"os"
)

var goroutineLocalEnv = routine.NewThreadLocal()

func functions(baseDir string) map[string]function.Function {
	r := hcl2template.Functions(baseDir)
	r["compliment"] = complimentFunction()
	r["env"] = envFunction()
	r["tostring"] = MakeToFunc(cty.String)
	r["tonumber"] = MakeToFunc(cty.Number)
	r["tobool"] = MakeToFunc(cty.Bool)
	r["toset"] = MakeToFunc(cty.Set(cty.DynamicPseudoType))
	r["tolist"] = MakeToFunc(cty.List(cty.DynamicPseudoType))
	r["tomap"] = MakeToFunc(cty.Map(cty.DynamicPseudoType))
	return r
}

func complimentFunction() function.Function {
	return function.New(&function.Spec{
		Description: "Return the compliment of list1 and all otherLists.",
		Params: []function.Parameter{
			{
				Name:             "list1",
				Description:      "the first list, will return all elements that in this list but not in any of other lists.",
				Type:             cty.Set(cty.DynamicPseudoType),
				AllowDynamicType: true,
			},
		},
		VarParam: &function.Parameter{
			Name:             "otherList",
			Description:      "other_list",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
		Type:         setOperationReturnType,
		RefineResult: refineNonNull,
		Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
			return s1.Subtract(s2)
		}, false),
	})
}

func envFunction() function.Function {
	return function.New(&function.Spec{
		Description: "Read environment variable, return empty string if the variable is not set.",
		Params: []function.Parameter{
			{
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
			envKey := key.AsString()
			localEnv := goroutineLocalEnv.Get()
			if localEnv != nil {
				if env, ok := (localEnv.(map[string]string))[envKey]; ok {
					return cty.StringVal(env), nil
				}
			}
			env := os.Getenv(envKey)
			return cty.StringVal(env), nil
		},
	})
}

func setOperationReturnType(args []cty.Value) (ret cty.Type, err error) {
	var etys []cty.Type
	for _, arg := range args {
		ty := arg.Type().ElementType()

		if arg.IsKnown() && arg.LengthInt() == 0 && ty.Equals(cty.DynamicPseudoType) {
			continue
		}

		etys = append(etys, ty)
	}

	if len(etys) == 0 {
		return cty.Set(cty.DynamicPseudoType), nil
	}

	newEty, _ := convert.UnifyUnsafe(etys)
	if newEty == cty.NilType {
		return cty.NilType, fmt.Errorf("given sets must all have compatible element types")
	}
	return cty.Set(newEty), nil
}

func refineNonNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}

func setOperationImpl(f func(s1, s2 cty.ValueSet) cty.ValueSet, allowUnknowns bool) function.ImplFunc {
	return func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		first := args[0]
		first, err = convert.Convert(first, retType)
		if err != nil {
			return cty.NilVal, function.NewArgError(0, err)
		}
		if !allowUnknowns && !first.IsWhollyKnown() {
			// This set function can produce a correct result only when all
			// elements are known, because eventually knowing the unknown
			// values may cause the result to have fewer known elements, or
			// might cause a result with no unknown elements at all to become
			// one with a different length.
			return cty.UnknownVal(retType), nil
		}

		set := first.AsValueSet()
		for i, arg := range args[1:] {
			arg, err := convert.Convert(arg, retType)
			if err != nil {
				return cty.NilVal, function.NewArgError(i+1, err)
			}
			if !allowUnknowns && !arg.IsWhollyKnown() {
				// (For the same reason as we did this check for "first" above.)
				return cty.UnknownVal(retType), nil
			}

			argSet := arg.AsValueSet()
			set = f(set, argSet)
		}
		return cty.SetValFromValueSet(set), nil
	}
}
