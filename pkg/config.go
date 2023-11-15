package pkg

import (
	"context"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/lonegunmanb/hclfuncs"
	"path/filepath"
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
		Functions: hclfuncs.Functions(c.basedir),
		Variables: map[string]cty.Value{
			"data": Values(c.DataSources),
			"rule": Values(c.Rules),
		},
	}
}

func (c *Config) parseFunc(expectedBlockType string, factories map[string]blockConstructor, blockRegisterFunc func(*Config, block)) func(*hclsyntax.Block) error {
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
		b := f(c, hb)
		blockRegisterFunc(c, b)
		err := eval(b)
		if err != nil {
			return fmt.Errorf("%s.%s.%s(%s) eval error: %+v", expectedBlockType, b.Type(), b.Name(), hb.Range().String(), err)
		}
		return nil
	}
}

func ParseConfig(baseDir, cfgDir string, ctx context.Context) (*Config, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	config := &Config{
		basedir: baseDir,
		ctx:     ctx,
	}

	var err error
	blocks, err := config.loadHclBlocks(cfgDir)
	if err != nil {
		return nil, err
	}

	// If there's dag error, return dag error first.
	config.dag, err = config.walkDag(blocks)
	if err != nil {
		return nil, err
	}

	parseErr := config.parseBlocks(dataParser, blocks)
	if parseErr != nil {
		return nil, parseErr
	}
	err = config.loadAllDataSources()
	if err != nil {
		return nil, err
	}

	parseErr = config.parseBlocks(ruleParser, blocks)
	if parseErr != nil {
		return nil, parseErr
	}
	parseErr = config.parseBlocks(fixParser, blocks)
	if parseErr != nil {
		return nil, parseErr
	}

	return config, parseErr
}

func (c *Config) parseBlocks(eval parserFactory, blocks []*hclsyntax.Block) error {
	var err error
	for _, b := range blocks {
		parseError := eval(c)(b)
		if parseError != nil {
			err = multierror.Append(err, parseError)
		}
	}
	return err
}

func (c *Config) loadHclBlocks(dir string) (hclsyntax.Blocks, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
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

func (c *Config) Plan() (*Plan, error) {
	var wg sync.WaitGroup
	plan := newPlan()
	errCh := make(chan error, len(c.Rules))

	// eval all rules
	for _, rule := range c.Rules {
		wg.Add(1)
		go func(rule Rule) {
			defer wg.Done()
			if err := eval(rule); err != nil {
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
			fr := &FailedRule{
				Rule:       rule,
				CheckError: checkErr,
			}
			plan.addRule(fr)

			// Find fixes for this rule
			for _, fix := range c.Fixes {
				refresh(fix)
				if linq.From(fix.GetRuleIds()).Contains(rule.Id()) {
					plan.addFix(fix)
				}
			}
		}(rule)
	}

	wg.Wait()
	close(errCh)

	err := readError(errCh)
	if err != nil {
		return nil, fmt.Errorf("the following blocks throw errors: %+v", err)
	}

	return plan, nil
}

func (c *Config) loadAllDataSources() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(c.DataSources))
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
			if err := eval(data); err != nil {
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
		return fmt.Errorf("the following blocks throw errors: %+v", err)
	}
	return nil
}

func (c *Config) walkDag(blocks hclsyntax.Blocks) (*dag.DAG, error) {
	g := dag.NewDAG()
	var walkErr error
	for _, b := range blocks {
		err := g.AddVertexByID(blockAddress(b), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.Body, dagWalker{dag: g, rootBlock: b})
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return g, walkErr
}

func readError(errors chan error) error {
	var err error
	for e := range errors {
		err = multierror.Append(err, e)
	}
	return err
}
