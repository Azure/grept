package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestFileHashRule_Check(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
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
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
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

func TestFileHashRule_HashMismatchFilesShouldBeExported(t *testing.T) {
	fs := afero.NewMemMapFs()
	filename := "/example/sub1/testfile.txt"
	_ = afero.WriteFile(fs, filename, []byte("test content"), 0644)
	rule := &FileHashRule{
		BaseRule: &BaseRule{
			c: &Config{},
		},
		Glob: "/example/*/testfile.txt",
		Hash: "non-matching-hash", // MD5 hash that doesn't match "test content"
	}
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()

	checkErr, runtimeErr := rule.Check()
	assert.Nil(t, runtimeErr)
	assert.NotNil(t, checkErr)
	assert.Contains(t, rule.HashMismatchFiles, filepath.FromSlash(filename))
}

func TestFileHashRule_FailOnHashMismatch(t *testing.T) {
	// Create a temporary file system
	fs := afero.NewMemMapFs()

	expectedContent := "hello"
	expectedHash := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // SHA256 of "hello"
	_ = afero.WriteFile(fs, "/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example/sub1/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example/sub2/testfile", []byte("world"), 0644)
	_ = afero.WriteFile(fs, "/example2/sub1/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "/example2/sub2/testfile", []byte(expectedContent), 0644)

	// Set the file system factory to use the temporary file system
	FsFactory = func() afero.Fs {
		return fs
	}

	tests := []struct {
		name      string
		rule      *FileHashRule
		wantErr   bool
		wantPaths []string
	}{
		{
			name: "FailOnHashMismatch is false, file content matches hash",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
		{
			name: "FailOnHashMismatch is false, file content does not match hash",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/testfile",
				Hash:               "incorrecthash",
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   true,
			wantPaths: []string{"/testfile"},
		},
		{
			name: "FailOnHashMismatch is false, one file content matches hash",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/example/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{"/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is true, file content does not match hash",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/example/*/testfile",
				Hash:               "incorrecthash",
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   true,
			wantPaths: []string{"/example/sub1/testfile", "/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is true, file content matches hash exits, but still got file hash that mismatch",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/example/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   true,
			wantPaths: []string{"/example/sub2/testfile"},
		},
		{
			name: "FailOnHashMismatch is false, all files match",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/example2/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: false,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
		{
			name: "FailOnHashMismatch is false, all files match",
			rule: &FileHashRule{
				BaseRule:           &BaseRule{},
				Glob:               "/example2/*/testfile",
				Hash:               expectedHash,
				Algorithm:          "sha256",
				FailOnHashMismatch: true,
			},
			wantErr:   false,
			wantPaths: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkErr, runtimeErr := tt.rule.Check()
			if runtimeErr != nil {
				t.Errorf("FileHashRule.Check() runtime error = %+v", runtimeErr)
			}
			if (checkErr != nil) != tt.wantErr {
				t.Errorf("FileHashRule.Check() error = %+v, wantErr %+v", checkErr, tt.wantErr)
			}
			var expectedPaths []string
			for i := 0; i < len(tt.rule.HashMismatchFiles); i++ {
				expectedPaths = append(expectedPaths, filepath.FromSlash(tt.rule.HashMismatchFiles[i]))
			}
			for _, path := range tt.wantPaths {
				assert.Contains(t, expectedPaths, filepath.FromSlash(path))
			}
		})
	}
}