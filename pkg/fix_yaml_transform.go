package pkg

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	yaml "github.com/lonegunmanb/atomatt-yaml"
	yamled "github.com/lonegunmanb/go-yaml-edit"
	"github.com/lonegunmanb/go-yaml-edit/splice"
	yptr "github.com/lonegunmanb/yaml-jsonpointer"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/text/transform"
)

var _ Fix = &YamlTransformFix{}

type YamlTransformFix struct {
	*BaseBlock
	baseFix
	RuleIds   []string        `hcl:"rule_ids" json:"rule_ids"`
	FilePath  string          `hcl:"file_path" json:"file_path" validate:"endswith=.yaml|endswith=.yml"`
	Transform []YamlTransform `hcl:"transform,block"`
}

type YamlTransform struct {
	YamlPath    string `hcl:"yaml_path" json:"yaml_path"`
	StringValue string `hcl:"string_value" json:"string_value"`
}

func (y *YamlTransformFix) Type() string {
	return "yaml_transform"
}

func (y *YamlTransformFix) Values() map[string]cty.Value {
	var transforms []cty.Value
	for _, t := range y.Transform {
		transforms = append(transforms, cty.ObjectVal(map[string]cty.Value{
			"yaml_path":    ToCtyValue(t.YamlPath),
			"string_value": ToCtyValue(t.StringValue),
		}))
	}
	return map[string]cty.Value{
		"rule_id":   ToCtyValue(y.RuleIds),
		"file_path": ToCtyValue(y.FilePath),
		"transform": cty.ListVal(transforms),
	}
}

func (y *YamlTransformFix) Apply() error {
	fs := FsFactory()
	yf, err := afero.ReadFile(fs, y.FilePath)
	if err != nil {
		return fmt.Errorf("error on reading yaml file %s, %+v fix.%s.%s %s", y.FilePath, err, y.Type(), y.Name(), y.HclSyntaxBlock().Range().String())
	}
	root := new(yaml.Node)
	err = yaml.Unmarshal(yf, root)
	if err != nil {
		return fmt.Errorf("error on ummarshal yaml file %s, %+v fix.%s.%s %s", y.FilePath, err, y.Type(), y.Name(), y.HclSyntaxBlock().Range().String())
	}
	var ops []splice.Op
	for _, t := range y.Transform {
		target, findErr := yptr.Find(root, t.YamlPath)
		if findErr != nil {
			err = multierror.Append(err, findErr)
			continue
		}
		ops = append(ops, yamled.Node(target).With(t.StringValue))
	}
	if err != nil {
		return fmt.Errorf("error on finding yaml node, %+v fix.%s.%s %s", err, y.Type(), y.Name(), y.HclSyntaxBlock().Range().String())
	}
	out, _, err := transform.Bytes(yamled.T(ops...), yf)
	if err != nil {
		return fmt.Errorf("error on transforming yaml node, %+v fix.%s.%s %s", err, y.Type(), y.Name(), y.HclSyntaxBlock().Range().String())
	}
	err = afero.WriteFile(fs, y.FilePath, out, 0600)
	if err != nil {
		return fmt.Errorf("error on writing yaml file %s, %+v fix.%s.%s %s", y.FilePath, err, y.Type(), y.Name(), y.HclSyntaxBlock().Range().String())
	}
	return nil
}

func (y *YamlTransformFix) GetRuleIds() []string {
	return y.RuleIds
}
