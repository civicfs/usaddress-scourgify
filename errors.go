package scourgify

import "fmt"

// AddressNormalizationError is the base error type for normalization errors
type AddressNormalizationError struct {
	Title   string
	Message string
	Details []interface{}
}

func (e *AddressNormalizationError) Error() string {
	if e.Title == "" && e.Message == "" {
		return "address normalization error"
	}
	msg := fmt.Sprintf("%s: %s", e.Title, e.Message)
	if len(e.Details) > 0 {
		msg += fmt.Sprintf(", %v", e.Details)
	}
	return msg
}

// AmbiguousAddressError indicates an error from ambiguous addresses or address parts
type AmbiguousAddressError struct {
	AddressNormalizationError
}

func NewAmbiguousAddressError(message string, details ...interface{}) *AmbiguousAddressError {
	title := "AMBIGUOUS ADDRESS"
	if message == "" {
		message = "This address contains ambiguous elements."
	}
	return &AmbiguousAddressError{
		AddressNormalizationError: AddressNormalizationError{
			Title:   title,
			Message: message,
			Details: details,
		},
	}
}

// UnParseableAddressError indicates an error from addresses that cannot be parsed
type UnParseableAddressError struct {
	AddressNormalizationError
}

func NewUnParseableAddressError(message string, details ...interface{}) *UnParseableAddressError {
	title := "UNPARSEABLE ADDRESS"
	if message == "" {
		message = "Unable to break this address into its component parts"
	}
	return &UnParseableAddressError{
		AddressNormalizationError: AddressNormalizationError{
			Title:   title,
			Message: message,
			Details: details,
		},
	}
}

// IncompleteAddressError indicates error from addresses that don't have enough data
type IncompleteAddressError struct {
	AddressNormalizationError
}

func NewIncompleteAddressError(message string, details ...interface{}) *IncompleteAddressError{
	title := "INCOMPLETE ADDRESS"
	if message == "" {
		message = "This address is missing one or more required elements"
	}
	return &IncompleteAddressError{
		AddressNormalizationError: AddressNormalizationError{
			Title:   title,
			Message: message,
			Details: details,
		},
	}
}

// AddressValidationError indicates address elements that don't meet format standards
type AddressValidationError struct {
	AddressNormalizationError
}

func NewAddressValidationError(message string, details ...interface{}) *AddressValidationError {
	title := "ADDRESS FORMAT VALIDATION"
	if message == "" {
		message = "Address contains invalid formatting"
	}
	return &AddressValidationError{
		AddressNormalizationError: AddressNormalizationError{
			Title:   title,
			Message: message,
			Details: details,
		},
	}
}
