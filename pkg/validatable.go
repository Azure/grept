package pkg

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

var validate = validator.New()

type Validatable interface {
	Validate() error
}

func ValidateConflictWith(fl validator.FieldLevel) bool {
	conflictWith := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	for _, anotherField := range conflictWith {
		field := parentStruct.FieldByName(anotherField)
		valid := field.IsValid()
		zero := field.IsZero()
		if valid && !zero {
			return false
		}
	}
	return true
}
