package pkg

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"runtime"
	"testing"
)

type localExecFixSuite struct {
	suite.Suite
}

func TestLocalExecFixSuite(t *testing.T) {
	suite.Run(t, new(localExecFixSuite))
}

func (s *localExecFixSuite) TestLocalExecShell_CommandValidate() {
	cases := []struct {
		desc      string
		f         *LocalShellFix
		wantError bool
	}{
		{
			desc: "Inlines only",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
			},
			wantError: false,
		},
		{
			desc: "Script only",
			f: &LocalShellFix{
				Inlines: nil,
				Script:  "./bash.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript only",
			f: &LocalShellFix{
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript http only",
			f: &LocalShellFix{
				RemoteScript: "http://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "non-http RemoteScript",
			f: &LocalShellFix{
				RemoteScript: "s3::https://s3.amazonaws.com/bucket/foo",
			},
			wantError: true,
		},
		{
			desc: "Inlines&Script",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
				Script:  "./bash.sh",
			},
			wantError: true,
		},
		{
			desc: "Script&RemoteScrip",
			f: &LocalShellFix{
				Inlines:      nil,
				Script:       "./bash.sh",
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc: "Inlines&RemoteScript",
			f: &LocalShellFix{
				Inlines:      []string{"echo hello"},
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc:      "No command",
			f:         &LocalShellFix{},
			wantError: true,
		},
		{
			desc: "InlineShebang Only",
			f: &LocalShellFix{
				Script:        "/test.sh",
				InlineShebang: "#!/bin/sh",
			},
			wantError: true,
		},
		{
			desc: "InlineShebang with inlines",
			f: &LocalShellFix{
				InlineShebang: "#!/bin/sh",
				Inlines: []string{
					"echo hello",
				},
			},
			wantError: false,
		},
		{
			desc: "Invalid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"Linux", // should be linux
				},
			},
			wantError: true,
		},
		{
			desc: "Valid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"linux",
					"windows",
					"darwin",
				},
			},
			wantError: false,
		},
	}
	for _, c := range cases {
		s.Run(c.desc, func() {
			err := validate.Struct(*c.f)
			if c.wantError {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})
	}
}

func (s *localExecFixSuite) TestLocalExecShell_ShouldReturnDirectlyIfOnlyOnMismatch() {
	quit := false
	stub := gostub.Stub(&stopByOnlyOnStub, func() {
		quit = true
	})
	defer stub.Reset()

	onlyOn := []string{"linux", "windows", "darwin"}
	linq.From(onlyOn).Where(func(i interface{}) bool {
		return i.(string) != runtime.GOOS // exclude the current os
	}).ToSlice(&onlyOn)

	sut := &LocalShellFix{
		Inlines: []string{"echo hello"},
		OnlyOn:  onlyOn,
	}
	err := sut.ApplyFix()
	assert.NoError(s.T(), err)
	assert.True(s.T(), quit)
}

func TestLocalShellFix_ApplyFix_inlines(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &Config{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		Inlines:        []string{`echo "Hello, World!"`},
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err := fix.ApplyFix()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// Check that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}
