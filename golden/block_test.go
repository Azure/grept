package golden

import (
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
	"math/big"
	"testing"
)

type TestData interface {
	PlanBlock
	Data()
}

type BaseData struct{}

func (db *BaseData) BlockType() string {
	return "data"
}

func (db *BaseData) Data() {}

func (db *BaseData) AddressLength() int { return 3 }

func (db *BaseData) CanExecutePrePlan() bool {
	return false
}

type TestResource interface {
	PlanBlock
	ApplyBlock
	Resource()
}

type BaseResource struct{}

func (rb *BaseResource) BlockType() string {
	return "resource"
}

func (rb *BaseResource) Resource() {}

func (rb *BaseResource) AddressLength() int {
	return 3
}

func (rb *BaseResource) CanExecutePrePlan() bool {
	return false
}

var _ TestData = &DummyData{}

type DummyData struct {
	*BaseData
	*BaseBlock
	Tags                      map[string]string `json:"data" hcl:"data,optional"`
	AttributeWithDefaultValue string            `json:"attribute" hcl:"attribute,optional" default:"default_value"`
}

func (d *DummyData) Type() string {
	return "dummy"
}

func (d *DummyData) ExecuteDuringPlan() error {
	return nil
}

var _ TestResource = &DummyResource{}

type DummyResource struct {
	*BaseBlock
	*BaseResource
	Tags map[string]string `json:"tags" hcl:"tags,optional"`
}

func (d *DummyResource) Type() string {
	return "dummy"
}

func (d *DummyResource) ExecuteDuringPlan() error {
	return nil
}

func (d *DummyResource) Apply() error {
	return nil
}

func Test_LocalBlocksValueShouldBeAFlattenObject(t *testing.T) {
	numberVal := cty.NumberVal(big.NewFloat(1))
	stringVal := cty.StringVal("hello world")
	locals := []Block{
		&LocalBlock{
			BaseBlock: &BaseBlock{
				name: "number_value",
			},
			LocalValue: numberVal,
		},
		&LocalBlock{
			BaseBlock: &BaseBlock{
				name: "string_value",
			},
			LocalValue: stringVal,
		},
	}

	values := SingleValues(castBlock[SingleValueBlock](locals))
	assert.True(t, AreCtyValuesEqual(numberVal, values.GetAttr("number_value")))
	assert.True(t, AreCtyValuesEqual(stringVal, values.GetAttr("string_value")))
}
