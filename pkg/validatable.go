package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/go-playground/validator/v10"
	"strings"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func registerValidator() {
	_ = validate.RegisterValidation("conflict_with", validateConflictWith)
	_ = validate.RegisterValidation("at_least_one_of", validateAtLeastOneOf)
	_ = validate.RegisterValidation("required_with", validateRequiredWith)
	_ = validate.RegisterValidation("all_string_in_slice", validateAllStringInSlice)
}

type Validatable interface {
	Validate() error
}

func validateAllStringInSlice(fl validator.FieldLevel) bool {
	candidatesQuery := linq.From(strings.Split(fl.Param(), " "))
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	values, ok := thisField.Interface().([]string)
	if !ok {
		return false
	}
	valuesQuery := linq.From(values)

	return !valuesQuery.Except(candidatesQuery).Any()
}

func validateConflictWith(fl validator.FieldLevel) bool {
	conflictWith := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	for _, anotherField := range conflictWith {
		field := parentStruct.FieldByName(anotherField)
		if field.IsValid() && !field.IsZero() {
			return false
		}
	}
	return true
}

func validateRequiredWith(fl validator.FieldLevel) bool {
	requiredWith := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	for _, anotherField := range requiredWith {
		field := parentStruct.FieldByName(anotherField)
		if !field.IsValid() || field.IsZero() {
			return false
		}
	}
	return true
}

func validateAtLeastOneOf(fl validator.FieldLevel) bool {
	atLeastOneOf := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	for _, fieldName := range atLeastOneOf {
		field := parentStruct.FieldByName(fieldName)
		if field.IsValid() && !field.IsZero() {
			return true
		}
	}
	return false
}
