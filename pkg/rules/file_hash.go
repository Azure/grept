package rules

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"

	"github.com/spf13/afero"
)

var _ Rule = &FileHashRule{}

type FileHashRule struct {
	fs        afero.Fs
	Glob      string `hcl:"glob,attr"`
	Hash      string `hcl:"hash,attr"`
	Algorithm string `hcl:"algorithm,attr"`
}

func (fhr *FileHashRule) Register(name string, factory func() Rule) {
	register("file_hash", func() Rule {
		return &FileHashRule{}
	})
}

func (fhr *FileHashRule) Check() error {
	// Use Glob to find files matching the path pattern
	files, err := afero.Glob(fhr.fs, fhr.Glob)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no files match path pattern: %s", fhr.Glob)
	}

	for _, file := range files {
		fileData, err := afero.ReadFile(fhr.fs, file)
		if err != nil {
			return err
		}

		// Calculate the hash of the file data
		var h hash.Hash
		switch fhr.Algorithm {
		case "md5":
			h = md5.New()
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		case "sha1":
			fallthrough
		default: // Default to sha1
			h = sha1.New()
		}
		h.Write(fileData)
		computedHash := fmt.Sprintf("%x", h.Sum(nil))

		if computedHash == fhr.Hash {
			return nil
		}
	}

	return fmt.Errorf("no file with glob %s and hash %s found", fhr.Glob, fhr.Hash)
}

func (fhr *FileHashRule) Validate() error {
	if fhr.Glob == "" {
		return fmt.Errorf("glob is required")
	}

	if fhr.Hash == "" {
		return fmt.Errorf("hash is required")
	}

	if fhr.Algorithm != "" {
		switch fhr.Algorithm {
		case "md5", "sha1", "sha256", "sha512":
			// valid
		default:
			return fmt.Errorf("invalid algorithm: %s", fhr.Algorithm)
		}
	}

	return nil
}
