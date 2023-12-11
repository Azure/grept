package pkg

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ahmetb/go-linq/v3"
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
	"github.com/lonegunmanb/hclfuncs"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var validBlockTypes sets.Set = hashset.New("data", "rule", "fix", "local")

type Config struct {
	ctx            context.Context
	basedir        string
	blockOperators map[string]*BlocksOperator
	dag            *dag.DAG
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
	if ctx == nil {
		ctx = context.Background()
	}
	config := newEmptyConfig()
	config.basedir = baseDir
	config.ctx = ctx

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
		t := b.BlockType()
		config.blockOperators[t].addBlock(b)
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

func (c *Config) runDag(onReady func(*Config, block)) error {
	c.resetWg()
	for _, b := range c.blocks() {
		b.setOnReady(onReady)
	}
	c.execErrChan = make(chan error, c.blocksCount())
	for _, n := range c.dag.GetRoots() {
		b := n.(block)
		go func() {
			onReady(c, b)
		}()
	}

	for _, operator := range c.blockOperators {
		operator.wg.Wait()
	}
	close(c.execErrChan)
	err := readError(c.execErrChan)
	if err != nil {
		return fmt.Errorf("the following blocks throw errors: %+v", err)
	}
	return nil
}

func (c *Config) planBlock(b block, errCh chan error) {
	if linq.From(b.getUpstreams()).AnyWith(func(i interface{}) bool {
		return !i.(block).getExecSuccess()
	}) {
		c.blockOperators[b.BlockType()].notifyOnExecuted(b, false)
		return
	}
	decodeErr := decode(b)
	if decodeErr != nil {
		errCh <- fmt.Errorf("%s.%s.%s(%s) decode error: %+v", b.Type(), b.Type(), b.Name(), b.HclSyntaxBlock().Range().String(), decodeErr)
		c.blockOperators[b.BlockType()].notifyOnExecuted(b, false)
		return
	}
	if v, ok := b.(Validatable); ok {
		if err := v.Validate(); err != nil {
			errCh <- fmt.Errorf("%s.%s.%s is not valid: %s", b.BlockType(), b.Type(), b.Name(), err.Error())
			c.blockOperators[b.BlockType()].notifyOnExecuted(b, false)
			return
		}
	}
	pa, ok := b.(planAction)
	if ok {
		execErr := pa.ExecuteDuringPlan()
		if execErr != nil {
			errCh <- fmt.Errorf("%s.%s.%s(%s) exec error: %+v", b.Type(), b.Type(), b.Name(), b.HclSyntaxBlock().Range().String(), execErr)
			c.blockOperators[b.BlockType()].notifyOnExecuted(b, false)
			return
		}
	}
	c.blockOperators[b.BlockType()].notifyOnExecuted(b, true)
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
			var bs []*hclsyntax.Block = readRawHclBlock(b)
			blocks = append(blocks, bs...)
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
	for _, v := range c.dag.GetVertices() {
		blocks = append(blocks, v.(block))
	}
	return blocks
}

func (c *Config) resetWg() {
	for _, o := range c.blockOperators {
		o.resetWg()
	}
}
