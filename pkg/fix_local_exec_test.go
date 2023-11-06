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

func (s *localExecFixSuite) TestLocalExec_CommandConflictWith() {
	cases := []struct {
		desc      string
		f         *LocalExecFix
		wantError bool
	}{
		{
			desc: "Inlines only",
			f: &LocalExecFix{
				Inlines: []string{"echo hello"},
			},
			wantError: false,
		},
		{
			desc: "Script only",
			f: &LocalExecFix{
				Inlines: nil,
				Script:  "./bash.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript only",
			f: &LocalExecFix{
				Inlines:      nil,
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "Inlines&Script",
			f: &LocalExecFix{
				Inlines: []string{"echo hello"},
				Script:  "./bash.sh",
			},
			wantError: true,
		},
		{
			desc: "Script&RemoteScrip",
			f: &LocalExecFix{
				Inlines:      nil,
				Script:       "./bash.sh",
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc: "RemoteScript only",
			f: &LocalExecFix{
				Inlines:      []string{"echo hello"},
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
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
