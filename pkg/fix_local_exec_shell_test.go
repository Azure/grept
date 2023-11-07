package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type localExecFixSuite struct {
	suite.Suite
}

func TestLocalExecFixSuite(t *testing.T) {
	suite.Run(t, new(localExecFixSuite))
}

func (s *localExecFixSuite) TestLocalExec_CommandValidate() {
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
