package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/prashantv/gostub"
	"testing"

	"github.com/spf13/afero"
)

func TestFileHashRule_Check(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&fsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	// Write some test files
	filePaths := []string{"./file1.txt", "./file2.txt", "./file3.txt", "./pkg/sub/subfile1.txt"}
	fileContents := []string{"hello", "world", "golang", "world"}
	for i, filePath := range filePaths {
		_ = afero.WriteFile(fs, filePath, []byte(fileContents[i]), 0644)
	}

	// Calculate the md5 hash of "world"
	h := md5.New()
	h.Write([]byte("world"))
	expectedHash := h.Sum(nil)

	tests := []struct {
		name      string
		rule      *FileHashRule
		wantError bool
	}{
		{
			name: "matching file found",
			rule: &FileHashRule{
				Glob:      "file*.txt",
				Hash:      fmt.Sprintf("%x", expectedHash),
				Algorithm: "md5",
			},
			wantError: false,
		},
		{
			name: "no matching file found",
			rule: &FileHashRule{
				Glob:      "file*.txt",
				Hash:      fmt.Sprintf("%x", sha1.Sum([]byte("world1"))),
				Algorithm: "sha1",
			},
			wantError: true,
		},
		{
			name: "no matching glob pattern",
			rule: &FileHashRule{
				Glob:      "nofile*.txt",
				Hash:      fmt.Sprintf("%x", expectedHash),
				Algorithm: "md5",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkError, _ := tt.rule.Check()
			if (checkError != nil) != tt.wantError {
				t.Errorf("FileHashRule.Check() error = %v, wantError %v", checkError, tt.wantError)
			}
		})
	}
}

func TestFileHashRule_Validate(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&fsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	tests := []struct {
		name      string
		rule      *FileHashRule
		wantError bool
	}{
		{
			name: "valid rule",
			rule: &FileHashRule{
				Glob:      "/file*.txt",
				Hash:      "abc123",
				Algorithm: "md5",
			},
			wantError: false,
		},
		{
			name: "missing glob",
			rule: &FileHashRule{
				Hash:      "abc123",
				Algorithm: "md5",
			},
			wantError: true,
		},
		{
			name: "missing hash",
			rule: &FileHashRule{
				Glob:      "/file*.txt",
				Algorithm: "md5",
			},
			wantError: true,
		},
		{
			name: "invalid algorithm",
			rule: &FileHashRule{
				Glob:      "/file*.txt",
				Hash:      "abc123",
				Algorithm: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("FileHashRule.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
