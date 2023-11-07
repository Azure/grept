package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
		f         *LocalExecShellFix
		wantError bool
	}{
		{
			desc: "Inlines only",
			f: &LocalExecShellFix{
				Inlines: []string{"echo hello"},
			},
			wantError: false,
		},
		{
			desc: "Script only",
			f: &LocalExecShellFix{
				Inlines: nil,
				Script:  "./bash.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript only",
			f: &LocalExecShellFix{
				Inlines:      nil,
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "Inlines&Script",
			f: &LocalExecShellFix{
				Inlines: []string{"echo hello"},
				Script:  "./bash.sh",
			},
			wantError: true,
		},
		{
			desc: "Script&RemoteScrip",
			f: &LocalExecShellFix{
				Inlines:      nil,
				Script:       "./bash.sh",
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc: "RemoteScript only",
			f: &LocalExecShellFix{
				Inlines:      []string{"echo hello"},
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc:      "No command",
			f:         &LocalExecShellFix{},
			wantError: true,
		},
		{
			desc: "InlineShebang Only",
			f: &LocalExecShellFix{
				Script:        "/test.sh",
				InlineShebang: "#!/bin/sh",
			},
			wantError: true,
		},
		{
			desc: "InlineShebang with inlines",
			f: &LocalExecShellFix{
				InlineShebang: "#!/bin/sh",
				Inlines: []string{
					"echo hello",
				},
			},
			wantError: false,
		},
		{
			desc: "Invalid only_on",
			f: &LocalExecShellFix{
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
			f: &LocalExecShellFix{
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

	sut := &LocalExecShellFix{
		Inlines: []string{"echo hello"},
		OnlyOn:  onlyOn,
	}
	err := sut.ApplyFix()
	assert.NoError(s.T(), err)
	assert.True(s.T(), quit)
}
