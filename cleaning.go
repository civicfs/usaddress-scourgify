package scourgify

import (
	"regexp"
	"strings"
	"unicode"
)

// AllowedChars contains character codes that are acceptable in addresses
// periods (in decimals), hyphens, /, and & are acceptable address components
var AllowedChars = []int{35, 38, 45, 46, 47} // '#', '&', '-', '.', '/'

// PreCleanExclude contains chars to exclude from PRE_CLEAN
// Don't remove ',', '(' or ')' in PRE_CLEAN
var PreCleanExclude = []int{40, 41, 44} // '(', ')', ','

// PreCleanAddrStr removes any known undesirable sub-strings and special characters.
//
// Cleaning should be enacted on an addr_str to remove known characters
// and phrases that might prevent address parsing from successfully working.
// Follows USPS pub 28 guidelines for undesirable special characters.
func PreCleanAddrStr(addrStr string, state string) string {
	// Replace any easily handled, undesirable sub-strings
	for oddity, replacement := range KnownOddities {
		if strings.Contains(addrStr, oddity) {
			addrStr = strings.ReplaceAll(addrStr, oddity, replacement)
		}
	}

	// Remove non-decimal point period chars
	if strings.Contains(addrStr, ".") {
		addrStr = CleanPeriodChar(addrStr)
	}

	addrStr = PreCleanDirectionals(addrStr)

	// Remove special characters per USPS pub 28
	excludeAll := append(AllowedChars, PreCleanExclude...)
	addrStr = CleanUpper(addrStr, excludeAll, false)

	// To prevent any potential confusion between CT = COURT v CT = Connecticut,
	// clean_ambiguous_street_types is not applied if state is CT.
	if state != "" {
		if _, exists := ProblemStTypeAbbrvs[state]; !exists {
			addrStr = CleanAmbiguousStreetTypes(addrStr)
		}
	}

	return addrStr
}

// CleanAmbiguousStreetTypes cleans street type abbreviations treated ambiguously.
//
// Some two char street type abbreviations (ie. CT) are treated as StateName
// by address parsers when address lines are parsed in isolation. To correct this,
// known problem abbreviations are converted to their whole word equivalent.
func CleanAmbiguousStreetTypes(addrStr string) string {
	if addrStr == "" {
		return addrStr
	}

	splitAddr := strings.Fields(addrStr)
	for key, value := range ProblemStTypeAbbrvs {
		for i, word := range splitAddr {
			if word == key {
				splitAddr[i] = value
				break
			}
		}
	}

	return strings.Join(splitAddr, " ")
}

// PostCleanAddrStr removes any special chars or extra white space remaining post-processing.
func PostCleanAddrStr(addrStr string) string {
	if addrStr == "" {
		return ""
	}

	return CleanUpper(addrStr, AllowedChars, false)
}

// CleanUpper returns text as upper case string and removes unwanted characters.
func CleanUpper(text string, exclude []int, stripSpaces bool) string {
	if text == "" {
		return ""
	}

	// Normalize unicode
	text = strings.ToUpper(text)

	// Create map of excluded characters for faster lookup
	excludeMap := make(map[rune]bool)
	for _, code := range exclude {
		excludeMap[rune(code)] = true
	}

	// Remove unwanted characters
	var result strings.Builder
	for _, r := range text {
		// Skip if it's a mark, symbol, or certain punctuation
		cat := unicode.In(r, unicode.M, unicode.S, unicode.C)
		isPunc := unicode.IsPunct(r)

		// Keep if it's excluded (allowed)
		if excludeMap[r] {
			result.WriteRune(r)
			continue
		}

		// Keep alphanumeric
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
			continue
		}

		// Keep spaces unless stripSpaces is true
		if unicode.IsSpace(r) && !stripSpaces {
			result.WriteRune(' ')
			continue
		}

		// Convert all dash-type characters to hyphen
		if unicode.In(r, unicode.Pd) {
			result.WriteRune('-')
			continue
		}

		// Skip marks, symbols, and punctuation not in exclude list
		if cat || isPunc {
			continue
		}

		// Keep anything else
		result.WriteRune(r)
	}

	// Remove extra spaces and return
	text = result.String()
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// CleanPeriodChar removes all period characters that are not decimal points.
func CleanPeriodChar(text string) string {
	// Remove periods not followed by digits
	pattern := regexp.MustCompile(`\.(?!\d)`)
	return pattern.ReplaceAllString(text, "")
}

// PreCleanDirectionals replaces any ambiguous directionals with their standard abbreviation.
//
// This helps ensure the directionals are correctly identified during address tagging,
// rather than being identified as part of the street name.
func PreCleanDirectionals(text string) string {
	text = strings.ToUpper(text)
	for direction, abbr := range AmbiguousDirectionals {
		text = strings.ReplaceAll(text, direction, abbr)
	}
	return text
}

// StripOccupancyType strips occupancy type (ie apt, unit, etc) from addr_line_2 string.
func StripOccupancyType(addrLine2 string) string {
	if addrLine2 == "" {
		return ""
	}

	addrLine2 = strings.TrimSpace(strings.ReplaceAll(addrLine2, "#", ""))
	addrLine2 = strings.ToUpper(addrLine2)

	// Simple implementation: remove known occupancy type abbreviations
	parts := strings.Fields(addrLine2)
	var result []string

	for _, part := range parts {
		// Check if this part is an occupancy type
		isOccType := false
		for key := range OccupancyTypeAbbreviations {
			if part == key {
				isOccType = true
				break
			}
		}
		for _, val := range OccupancyTypeAbbreviations {
			if part == val {
				isOccType = true
				break
			}
		}

		// Only keep parts that aren't occupancy types
		if !isOccType {
			result = append(result, part)
		}
	}

	return strings.Join(result, " ")
}
