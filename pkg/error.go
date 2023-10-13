package pkg

import (
	"fmt"
	"strings"
)

type blockError struct {
	BlockCategory string
	BlockType     string
	BlockName     string
	Err           error
}

func (pe blockError) Error() string {
	return fmt.Sprintf("%s.%s.%s error: %s", pe.BlockCategory, pe.BlockType, pe.BlockName, pe.Err.Error())
}

type blockErrors struct {
	Errors []error
}

func (pe *blockErrors) Error() string {
	errMsgs := make([]string, len(pe.Errors))
	for i, err := range pe.Errors {
		errMsgs[i] = err.Error()
	}
	return fmt.Sprintf("following blocks throw errors:\n\t%s", strings.Join(errMsgs, "\n\t"))
}

func (pe *blockErrors) Add(err error) *blockErrors {
	e := pe
	if pe == nil {
		e = &blockErrors{}
	}
	e.Errors = append(e.Errors, err)
	return e
}
