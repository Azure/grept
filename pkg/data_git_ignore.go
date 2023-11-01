package pkg

import (
	"bufio"
	"github.com/spf13/afero"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var _ Data = &GitIgnore{}

type GitIgnore struct {
	*BaseData
	Records []string
}

func (g *GitIgnore) Load() error {
	fs := FsFactory()
	gitIgnoreFile := "./.gitignore"
	exists, err := afero.Exists(fs, gitIgnoreFile)
	if err != nil {
		return err
	}
	if !exists {
		g.Records = []string{}
		return nil
	}
	f, err := fs.Open(gitIgnoreFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		g.Records = append(g.Records, line)
	}
	return scanner.Err()
}

func (g *GitIgnore) Type() string {
	return "git_ignore"
}

func (g *GitIgnore) Value() cty.Value {
	values := g.BaseValue()
	values["records"] = ToCtyValue(g.Records)
	return cty.ObjectVal(values)
}

func (g *GitIgnore) Eval(h *hclsyntax.Block) error {
	var err error
	if err = g.BaseData.Parse(h); err != nil {
		return err
	}
	diag := gohcl.DecodeBody(h.Body, g.EvalContext(), g)
	if diag.HasErrors() {
		return diag
	}
	return nil
}
