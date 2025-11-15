package scourgify

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidateAddressComponents validates non-null values for minimally viable address elements.
//
// All addresses should have at least an address_line_1 and a postal_code
// or a city and state.
func ValidateAddressComponents(addressDict map[string]string, strict bool) (map[string]string, error) {
	var locality bool

	if strict {
		locality = addressDict["postal_code"] != "" &&
			addressDict["city"] != "" &&
			addressDict["state"] != ""
	} else {
		locality = addressDict["postal_code"] != "" ||
			(addressDict["city"] != "" && addressDict["state"] != "")
	}

	if addressDict["address_line_1"] == "" {
		return nil, NewIncompleteAddressError("Address records must include Line 1 data.")
	}

	if !locality {
		msg := "Address records must contain a city, state, and postal_code."
		if !strict {
			msg = "Address records must contain a city and state, or a postal_code"
		}
		return nil, NewIncompleteAddressError(msg)
	}

	return addressDict, nil
}

// ValidateUSPostalCodeFormat validates postal code conforms to US five-digit Zip or Zip+4 standards.
func ValidateUSPostalCodeFormat(postalCode string, address interface{}) (string, error) {
	hasError := false
	msg := "US Postal Codes must conform to five-digit Zip or Zip+4 standards."

	postalCode = PostCleanAddrStr(postalCode)
	plusFourCode := strings.Split(postalCode, "-")

	// Check if all parts are numeric
	for _, code := range plusFourCode {
		if _, err := strconv.Atoi(code); err != nil {
			hasError = true
			break
		}
	}

	if !hasError {
		if strings.Contains(postalCode, "-") {
			if len(strings.ReplaceAll(postalCode, "-", "")) > 9 {
				hasError = true
			} else if len(plusFourCode) != 2 {
				hasError = true
			} else {
				// Zero-pad the zip and plus-4
				postalCode = fmt.Sprintf("%05s-%04s", plusFourCode[0], plusFourCode[1])
			}
		} else if len(postalCode) == 9 {
			postalCode = postalCode[:5] + "-" + postalCode[5:]
		} else if len(postalCode) > 5 {
			hasError = true
		} else {
			// Zero-pad the zip
			postalCode = fmt.Sprintf("%05s", postalCode)
		}
	}

	if hasError {
		return "", NewAddressValidationError(msg, address)
	}

	return postalCode, nil
}

// ValidateParensGroupsParsed validates any parenthesis segments have been successfully parsed.
//
// Assumes any parenthesis segments in original address string are either
// line 2 or ambiguous address elements. If any parenthesis segment remains
// in line1 after all other address processing has been applied,
// AmbiguousAddressError is raised.
func ValidateParensGroupsParsed(line1 string) (string, error) {
	pattern := regexp.MustCompile(`\((.+?)\)`)
	parenthesisGroups := pattern.FindAllString(line1, -1)

	if len(parenthesisGroups) > 0 {
		return "", NewAmbiguousAddressError("", line1)
	}

	return line1, nil
}
