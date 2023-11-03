package pkg

import (
	"bufio"
	"github.com/spf13/afero"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var _ Data = &GitIgnoreDatasource{}

type GitIgnoreDatasource struct {
	baseData
	Records []string
}

func (g *GitIgnoreDatasource) Load() error {
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
		line = strings.TrimFunc(line, func(r rune) bool {
			return r == '\t' || r == ' '
		})
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		g.Records = append(g.Records, line)
	}
	return scanner.Err()
}

func (g *GitIgnoreDatasource) Type() string {
	return "git_ignore"
}

func (g *GitIgnoreDatasource) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"records": ToCtyValue(g.Records),
	}
}

func (g *GitIgnoreDatasource) Eval(h *hclsyntax.Block) error {
	var err error
	if err = g.baseBlock.Parse(h); err != nil {
		return err
	}
	diag := gohcl.DecodeBody(h.Body, g.EvalContext(), g)
	if diag.HasErrors() {
		return diag
	}
	return nil
}