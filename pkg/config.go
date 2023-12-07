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

type Config struct {
	ctx           context.Context
	basedir       string
	DatasOperator *BlocksOperator
	RulesOperator *BlocksOperator
	FixesOperator *BlocksOperator
	dag           *dag.DAG
}

func (c *Config) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: hclfuncs.Functions(c.basedir),
		Variables: map[string]cty.Value{
			"data": Values(c.DatasOperator.blocks),
			"rule": Values(c.RulesOperator.blocks),
		},
	}
}

func (c *Config) parseFunc(expectedBlockType string, factories map[string]blockConstructor) func(*hclsyntax.Block) error {
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
		err := eval(b)
		if err != nil {
			return fmt.Errorf("%s.%s.%s(%s) eval error: %+v", expectedBlockType, b.Type(), b.Name(), hb.Range().String(), err)
		}
		return nil
	}
}

func wrapBlock(c *Config, hb *hclsyntax.Block) (block, error) {
	blockFactories := factories[hb.Type]
	blockType := ""
	if len(hb.Labels) > 0 {
		blockType = hb.Labels[0]
	}
	f, ok := blockFactories[blockType]
	if !ok {
		return nil, fmt.Errorf("unregistered %s: %s", hb.Type, blockType)
	}
	return f(c, hb), nil
}

func NewConfig(baseDir, cfgDir string, ctx context.Context) (*Config, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	config := &Config{
		basedir:       baseDir,
		ctx:           ctx,
		DatasOperator: &BlocksOperator{},
		RulesOperator: &BlocksOperator{},
		FixesOperator: &BlocksOperator{},
	}

	var err error
	hclBlocks, err := config.loadHclBlocks(cfgDir)
	if err != nil {
		return nil, err
	}

	var blocks []block
	for _, hb := range hclBlocks {
		b, wrapError := wrapBlock(config, hb)
		if wrapError != nil {
			err = multierror.Append(wrapError)
			continue
		}
		blocks = append(blocks, b)
		if r, isRule := b.(Rule); isRule {
			config.RulesOperator.addBlock(r)
		} else if d, isData := b.(Data); isData {
			config.DatasOperator.addBlock(d)
		} else if f, isFix := b.(Fix); isFix {
			config.FixesOperator.addBlock(f)
		}
	}
	if err != nil {
		return nil, err
	}

	// If there's dag error, return dag error first.
	config.dag, err = config.walkDag(blocks)
	if err != nil {
		return nil, err
	}
	return config, nil
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
	var hclBlocks hclsyntax.Blocks
	for _, v := range c.dag.GetVertices() {
		hclBlocks = append(hclBlocks, v.(block).HclSyntaxBlock())
	}
	var err error
	for _, d := range c.DatasOperator.blocks {
		evalErr := eval(d)
		if evalErr != nil {
			err = multierror.Append(err, fmt.Errorf("%s.%s.%s(%s) eval error: %+v", d.Type(), d.Type(), d.Name(), d.HclSyntaxBlock().Range().String(), evalErr))
		}
	}
	if err != nil {
		return nil, err
	}
	err = c.loadAllDataSources()
	if err != nil {
		return nil, err
	}

	for _, r := range c.RulesOperator.blocks {
		evalErr := eval(r)
		if evalErr != nil {
			err = multierror.Append(err, fmt.Errorf("%s.%s.%s(%s) eval error: %+v", r.Type(), r.Type(), r.Name(), r.HclSyntaxBlock().Range().String(), evalErr))
		}
	}
	if err != nil {
		return nil, err
	}

	for _, f := range c.FixesOperator.blocks {
		evalErr := eval(f)
		if evalErr != nil {
			err = multierror.Append(err, fmt.Errorf("%s.%s.%s(%s) eval error: %+v", f.Type(), f.Type(), f.Name(), f.HclSyntaxBlock().Range().String(), evalErr))
		}
	}
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	plan := newPlan()
	errCh := make(chan error, len(c.RulesOperator.blocks))

	// eval all rules
	for _, rule := range c.RulesOperator.blocks {
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
			for _, f := range c.FixesOperator.blocks {
				fix := f.(Fix)
				refresh(f)

				if linq.From(fix.GetRuleIds()).Contains(rule.Id()) {
					plan.addFix(fix)
				}
			}
		}(rule.(Rule))
	}

	wg.Wait()
	close(errCh)

	err = readError(errCh)
	if err != nil {
		return nil, fmt.Errorf("the following blocks throw errors: %+v", err)
	}

	return plan, nil
}

func (c *Config) loadAllDataSources() error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(c.DatasOperator.blocks))
	// Load all datasources
	for _, data := range c.DatasOperator.blocks {
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
		}(data.(Data))
	}
	wg.Wait()
	close(errCh)
	err := readError(errCh)
	if err != nil {
		return fmt.Errorf("the following blocks throw errors: %+v", err)
	}
	return nil
}

func (c *Config) walkDag(blocks []block) (*dag.DAG, error) {
	g := dag.NewDAG()
	var walkErr error
	for _, b := range blocks {
		err := g.AddVertexByID(blockAddress(b.HclSyntaxBlock()), b)
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

func readError(errors chan error) error {
	var err error
	for e := range errors {
		err = multierror.Append(err, e)
	}
	return err
}
