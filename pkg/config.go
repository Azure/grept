package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var validBlockTypes sets.Set = hashset.New("data", "rule", "fix")

type parserFactory func(c *Config) func(*hclsyntax.Block) error

var ruleParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("rule", ruleFactories, func(cc *Config, b block) {
		cc.Rules = append(cc.Rules, b.(Rule))
	})

}
var fixParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("fix", fixFactories, func(cc *Config, b block) {
		cc.Fixes = append(cc.Fixes, b.(Fix))
	})
}
var dataParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("data", datasourceFactories, func(cc *Config, b block) {
		cc.DataSources = append(cc.DataSources, b.(Data))
	})
}

type parsers []parserFactory

var blockParsers parsers = []parserFactory{
	dataParser,
	ruleParser,
	fixParser,
}

type Datas []Data
type Rules []Rule
type Fixes []Fix

type Config struct {
	ctx         context.Context
	basedir     string
	DataSources Datas
	Rules       Rules
	Fixes       Fixes
	dag         *dag.DAG
}

func (c *Config) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: functions(c.basedir),
		Variables: map[string]cty.Value{
			"data": Values(c.DataSources),
			"rule": Values(c.Rules),
		},
	}
}

func (c *Config) parseFunc(expectedBlockType string, factories map[string]func(*Config) block, blockRegisterFunc func(*Config, block)) func(*hclsyntax.Block) error {
	return func(hb *hclsyntax.Block) error {
		if hb.Type != expectedBlockType {
			return nil
		}
		if len(hb.Labels) != 2 {
			return fmt.Errorf("invalid labels for %s %s, expect labels with length 2 (%s)", expectedBlockType, concatLabels(hb.Labels), hb.Range().String())
		}
		t := hb.Labels[0]
		f, ok := factories[t]
		if !ok {
			return fmt.Errorf("unregistered %s: %s, %s", expectedBlockType, t, hb.Range().String())
		}
		b := f(c)
		blockRegisterFunc(c, b)
		err := b.Eval(hb)
		if err != nil {
			return fmt.Errorf("%s.%s.%s(%s) eval error: %+v", expectedBlockType, b.Type(), b.Name(), hb.Range().String(), err)
		}
		return nil
	}
}

func ParseConfig(dir string, ctx context.Context) (*Config, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	config := &Config{
		basedir: dir,
		ctx:     ctx,
	}

	var err error
	blocks, err := config.loadHclBlocks(dir)
	if err != nil {
		return nil, err
	}
	parseErr := config.parseBlocks(blocks)
	// If there's dag error, return dag error first.
	config.dag, err = config.walkDag()
	if err != nil {
		return nil, err
	}

	return config, parseErr
}

func (c *Config) parseBlocks(blocks []*hclsyntax.Block) error {
	var err error
	for _, eval := range blockParsers {
		for _, b := range blocks {
			parseError := eval(c)(b)
			if parseError != nil {
				err = multierror.Append(err, parseError)
			}
		}
	}
	return err
}

func (c *Config) loadHclBlocks(dir string) ([]*hclsyntax.Block, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, fmt.Sprintf(filepath.Join(dir, "*.grept.hcl")))
	if err != nil {
		return nil, err
	}

	var blocks []*hclsyntax.Block

	for _, filename := range matches {
		content, fsErr := afero.ReadFile(fs, filename)
		if fsErr != nil {
			err = multierror.Append(err, fsErr)
			continue
		}
		file, diag := hclsyntax.ParseConfig(content, filename, hcl.InitialPos)
		if diag.HasErrors() {
			err = multierror.Append(err, diag.Errs()...)
			continue
		}
		body := file.Body.(*hclsyntax.Body)
		for _, b := range body.Blocks {
			blocks = append(blocks, b)
		}
	}
	if err != nil {
		return nil, err
	}

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if !validBlockTypes.Contains(b.Type) {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
			continue
		}
	}
	return blocks, err
}

func (c *Config) Plan() (Plan, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(c.DataSources))
	plan := make(Plan)

	// Load all datasources
	for _, data := range c.DataSources {
		wg.Add(1)
		go func(data Data) {
			defer wg.Done()

			if v, ok := data.(Validatable); ok {
				if err := v.Validate(); err != nil {
					errCh <- fmt.Errorf("data.%s.%s is not valid: %s", data.Type(), data.Name(), err.Error())
					return
				}
			}
			if err := data.Eval(data.HclSyntaxBlock()); err != nil {
				errCh <- fmt.Errorf("data.%s.%s(%s) eval error: %+v", data.Type(), data.Name(), data.HclSyntaxBlock().Range().String(), err)
				return
			}
			if err := data.Load(); err != nil {
				errCh <- fmt.Errorf("data.%s.%s(%s) throws error: %s", data.Type(), data.Name(), data.HclSyntaxBlock().Range().String(), err.Error())
			}
		}(data)
	}
	wg.Wait()
	close(errCh)
	err := readError(errCh)
	if err != nil {
		return nil, fmt.Errorf("the following blocks throw errors: %+v", err)
	}

	errCh = make(chan error, len(c.Rules))

	// Eval all rules
	for _, rule := range c.Rules {
		wg.Add(1)
		go func(rule Rule) {
			defer wg.Done()
			if err := rule.Eval(rule.HclSyntaxBlock()); err != nil {
				errCh <- fmt.Errorf("rule.%s.%s(%s) eval error: %+v", rule.Type(), rule.Name(), rule.HclSyntaxBlock().Range().String(), err)
				return
			}
			checkErr, runtimeErr := rule.Check()
			if runtimeErr != nil {
				errCh <- runtimeErr
				return
			}
			if checkErr == nil {
				// This rule passes check, no need to fix it
				return
			}
			fr := &failedRule{
				Rule:       rule,
				CheckError: checkErr,
			}

			plan[fr] = make(Fixes, 0)
			// Find fixes for this rule
			for _, fix := range c.Fixes {
				refresh(fix)
				if fix.GetRuleId() == rule.Id() {
					plan[fr] = append(plan[fr], fix)
				}
			}
		}(rule)
	}

	wg.Wait()
	close(errCh)

	err = readError(errCh)
	if err != nil {
		return nil, fmt.Errorf("the following blocks throw errors: %+v", err)
	}

	return plan, nil
}

func (c *Config) walkDag() (*dag.DAG, error) {
	g := dag.NewDAG()
	var blocks []block
	for _, d := range c.DataSources {
		blocks = append(blocks, d)
	}
	for _, r := range c.Rules {
		blocks = append(blocks, r)
	}
	for _, f := range c.Fixes {
		blocks = append(blocks, f)
	}
	var walkErr error
	for _, b := range blocks {
		err := g.AddVertexByID(address(b), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.HclSyntaxBlock().Body, dagWalker{dag: g, rootBlock: b})
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return g, walkErr
}

func address(b block) string {
	sb := strings.Builder{}
	sb.WriteString(b.BlockType())
	sb.WriteString(".")
	if t := b.Type(); t != "" {
		sb.WriteString(t)
		sb.WriteString(".")
	}
	sb.WriteString(b.Name())
	return sb.String()
}

func readError(errors chan error) error {
	var err error
	for e := range errors {
		err = multierror.Append(err, e)
	}
	return err
}
