package pkg

import (
	"context"
	"fmt"
	"github.com/emirpasic/gods/queues/linkedlistqueue"
	"path/filepath"

	"github.com/ahmetb/go-linq/v3"
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/lonegunmanb/hclfuncs"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var validBlockTypes sets.Set = hashset.New("data", "rule", "fix", "local")

type Config struct {
	ctx            context.Context
	basedir        string
	blockOperators map[string]*BlocksOperator
	dag            *Dag
	execErrChan    chan error
}

func (c *Config) DatasOperator() *BlocksOperator {
	return c.blockOperators["data"]
}

func (c *Config) RulesOperator() *BlocksOperator {
	return c.blockOperators["rule"]
}

func (c *Config) FixesOperator() *BlocksOperator {
	return c.blockOperators["fix"]
}

func (c *Config) LocalsOperator() *BlocksOperator {
	return c.blockOperators["local"]
}

func (c *Config) DataBlocks() []Data {
	return contravariance[Data](c.DatasOperator().Blocks())
}

func (c *Config) RuleBlocks() []Rule {
	return contravariance[Rule](c.RulesOperator().Blocks())
}

func (c *Config) FixBlocks() []Fix {
	return contravariance[Fix](c.FixesOperator().Blocks())
}

func (c *Config) LocalBlocks() []Local {
	return contravariance[Local](c.LocalsOperator().Blocks())
}

func (c *Config) blocksCount() int {
	var cnt int
	for _, o := range c.blockOperators {
		cnt += o.blocksCount()
	}
	return cnt
}

func contravariance[T block](blocks []block) []T {
	var r []T
	for _, b := range blocks {
		r = append(r, b.(T))
	}
	return r
}

func (c *Config) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: hclfuncs.Functions(c.basedir),
		Variables: map[string]cty.Value{
			"data":  Values(c.DataBlocks()),
			"rule":  Values(c.RuleBlocks()),
			"local": LocalsValues(c.LocalBlocks()),
		},
	}
}

func newEmptyConfig() *Config {
	c := &Config{
		ctx: context.TODO(),
	}
	c.blockOperators = map[string]*BlocksOperator{
		"data":  NewBlocksOperator(c),
		"rule":  NewBlocksOperator(c),
		"fix":   NewBlocksOperator(c),
		"local": NewBlocksOperator(c),
	}
	return c
}

func NewConfig(baseDir, cfgDir string, ctx context.Context) (*Config, error) {
	var err error
	hclBlocks, err := loadHclBlocks(cfgDir)
	if err != nil {
		return nil, err
	}

	c, err := newConfig(baseDir, ctx, hclBlocks, err)
	if err != nil {
		return nil, err
	}
	expandedHclBlocks := make([]*hclBlock, 0)
	for _, b := range c.blocks() {
		hb := b.HclBlock()
		attr, ok := hb.Body.Attributes["for_each"]
		if !ok {
			expandedHclBlocks = append(expandedHclBlocks, newHclBlock(hb.Block, nil))
			continue
		}
		forEachValue, diag := attr.Expr.Value(c.EvalContext())
		if diag.HasErrors() {
			err = multierror.Append(err, diag)
			continue
		}
		if !forEachValue.CanIterateElements() {
			err = multierror.Append(err, fmt.Errorf("invalid `for_each`, except set or map: %s", attr.Range().String()))
			continue
		}
		iterator := forEachValue.ElementIterator()
		for iterator.Next() {
			key, value := iterator.Element()
			newBlock := newHclBlock(hb.Block, &forEach{key: key, value: value})
			expandedHclBlocks = append(expandedHclBlocks, newBlock)
		}
	}
	if len(expandedHclBlocks) != len(hclBlocks) {
		c, err = newConfig(baseDir, ctx, expandedHclBlocks, err)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func newConfig(baseDir string, ctx context.Context, hclBlocks []*hclBlock, err error) (*Config, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	config := newEmptyConfig()
	config.basedir = baseDir
	config.ctx = ctx

	var blocks []block
	for _, hb := range hclBlocks {
		b, wrapError := wrapBlock(config, hb)
		if wrapError != nil {
			err = multierror.Append(wrapError)
			continue
		}
		blocks = append(blocks, b)
		t := b.BlockType()
		config.blockOperators[t].addBlock(b)
	}
	if err != nil {
		return nil, err
	}
	err = config.runDag(prepare)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Plan() (*Plan, error) {
	err := c.runDag(plan)
	if err != nil {
		return nil, err
	}

	plan := newPlan()
	for _, rb := range c.RuleBlocks() {
		checkErr := rb.CheckError()
		if checkErr != nil {
			plan.addRule(&FailedRule{
				Rule:       rb,
				CheckError: checkErr,
			})
		}
		for _, fb := range c.FixBlocks() {
			if linq.From(fb.GetRuleIds()).Contains(rb.Id()) {
				plan.addFix(fb)
			}
		}
	}

	return plan, nil
}

func (c *Config) runDag(onReady func(*Config, *Dag, block) error) error {
	// If there's dag error, return dag error first.
	dag, err := newDag(c.blocks())
	if err != nil {
		return err
	}
	visited := hashset.New()
	pending := linkedlistqueue.New()
	for _, n := range dag.GetRoots() {
		pending.Enqueue(n.(block))
	}
	for !pending.Empty() {
		next, _ := pending.Dequeue()
		b := next.(block)
		ancestors, dagErr := dag.GetAncestors(blockAddress(b.HclBlock()))
		if dagErr != nil {
			return dagErr
		}
		ready := true
		for upstreamAddress, _ := range ancestors {
			if !visited.Contains(upstreamAddress) {
				ready = false
			}
		}
		if !ready {
			continue
		}
		if callbackErr := onReady(c, dag, b); callbackErr != nil {
			err = multierror.Append(err, callbackErr)
		}
		visited.Add(blockAddress(b.HclBlock()))
		children, dagErr := dag.GetChildren(blockAddress(b.HclBlock()))
		if dagErr != nil {
			return dagErr
		}
		for _, n := range children {
			pending.Enqueue(n)
		}
	}
	return err
}

func (c *Config) planBlock(b block) error {
	decodeErr := decode(b)
	if decodeErr != nil {
		return fmt.Errorf("%s.%s.%s(%s) decode error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), decodeErr)
	}
	if v, ok := b.(Validatable); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("%s.%s.%s is not valid: %s", b.BlockType(), b.Type(), b.Name(), err.Error())
		}
	}
	pa, ok := b.(planAction)
	if ok {
		execErr := pa.ExecuteDuringPlan()
		if execErr != nil {
			return fmt.Errorf("%s.%s.%s(%s) exec error: %+v", b.Type(), b.Type(), b.Name(), b.HclBlock().Range().String(), execErr)
		}
	}
	return nil
}

func readError(errors chan error) error {
	var err error
	for e := range errors {
		err = multierror.Append(err, e)
	}
	return err
}

func wrapBlock(c *Config, hb *hclBlock) (block, error) {
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

func loadHclBlocks(dir string) ([]*hclBlock, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
	}

	var blocks []*hclBlock

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
			var bs []*hclsyntax.Block = readRawHclBlock(b)
			for _, hb := range bs {
				blocks = append(blocks, newHclBlock(hb, nil))
			}
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

func readRawHclBlock(b *hclsyntax.Block) []*hclsyntax.Block {
	if b.Type != "locals" {
		return []*hclsyntax.Block{b}
	}
	var newBlocks []*hclsyntax.Block
	for _, attr := range b.Body.Attributes {
		newBlocks = append(newBlocks, &hclsyntax.Block{
			Type:   "local",
			Labels: []string{"", attr.Name},
			Body: &hclsyntax.Body{
				Attributes: map[string]*hclsyntax.Attribute{
					"value": {
						Name:        "value",
						Expr:        attr.Expr,
						SrcRange:    attr.SrcRange,
						NameRange:   attr.NameRange,
						EqualsRange: attr.EqualsRange,
					},
				},
				SrcRange: attr.NameRange,
				EndRange: attr.SrcRange,
			},
		})
	}
	return newBlocks
}

func (c *Config) blocks() []block {
	var blocks []block
	for _, o := range c.blockOperators {
		blocks = append(blocks, o.Blocks()...)
	}
	return blocks
}
