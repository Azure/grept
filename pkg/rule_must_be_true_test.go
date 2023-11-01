package pkg

import (
	"context"
	"fmt"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"testing"
)

func TestMustBeTrueRule_Check(t *testing.T) {
	tests := []struct {
		name        string
		rule        *MustBeTrueRule
		expectError bool
	}{
		{
			name: "Condition is true",
			rule: &MustBeTrueRule{
				BaseRule:     &BaseRule{},
				Condition:    true,
				ErrorMessage: "",
			},
			expectError: false,
		},
		{
			name: "Condition is false",
			rule: &MustBeTrueRule{
				BaseRule:     &BaseRule{},
				Condition:    false,
				ErrorMessage: "Test error message",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, runTimeErr := tt.rule.Check()
			require.NoError(t, runTimeErr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "assertion failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMustBeTrueRule_Eval(t *testing.T) {
	expectedContent := "hello world"
	temp, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	_, err = temp.WriteString(expectedContent)
	require.NoError(t, err)
	err = temp.Close()
	defer func() {
		s := temp.Name()
		err := os.Remove(s)
		print(err)
	}()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/test.grept.hcl", []byte(fmt.Sprintf(`
	rule "must_be_true" sample {
		condition = file("%s") == "hello world"
	}
`, filepath.ToSlash(temp.Name()))), 0644)
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()

	c, err := ParseConfig("/", context.TODO())
	require.NoError(t, err)
	plan, err := c.Plan()
	require.NoError(t, err)
	assert.Empty(t, plan)
}

func TestMustBeTrueRule_Value(t *testing.T) {
	mustBeTrueRule := &MustBeTrueRule{
		BaseRule:     &BaseRule{},
		Condition:    true,
		ErrorMessage: "Test error message",
	}

	value := make(map[string]cty.Value)
	mustBeTrueRule.SetValues(value)

	assert.Equal(t, map[string]cty.Value{
		"condition":     cty.BoolVal(mustBeTrueRule.Condition),
		"error_message": cty.StringVal(mustBeTrueRule.ErrorMessage),
	}, value)
}
