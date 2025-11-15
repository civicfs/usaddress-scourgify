package scourgify

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Address represents a normalized US address
type Address struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
}

// NormalizeAddressRecord normalizes an address string or map.
//
// Takes an address string or map with standard address fields, removes
// unacceptable special characters, extra spaces, predictable abnormal
// character sub-strings and phrases, abbreviates directional indicators
// and street types.
func NormalizeAddressRecord(address interface{}, strict bool, longHand bool) (*Address, error) {
	switch v := address.(type) {
	case string:
		return NormalizeAddrStr(v, "", "", "", "", longHand)
	case map[string]string:
		return NormalizeAddrDict(v, strict, longHand)
	default:
		return nil, fmt.Errorf("address must be a string or map[string]string")
	}
}

// NormalizeAddrStr normalizes a complete or partial address string.
func NormalizeAddrStr(addrStr, line2, city, state, zipcode string, longHand bool) (*Address, error) {
	// Normalize state first if provided
	normalizedState := NormalizeState(state)

	// Pre-clean the address string
	addrStr = PreCleanAddrStr(addrStr, normalizedState)

	// Simple parsing logic - split by comma to extract components
	parts := strings.Split(addrStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var line1 string
	if len(parts) > 0 {
		line1 = parts[0]
	}

	// Try to extract city, state, zip from remaining parts
	if city == "" && len(parts) > 1 {
		city = parts[1]
	}

	if normalizedState == "" && len(parts) > 2 {
		// Check if state and zip are in separate parts (e.g., "Portland, OR, 97203")
		if len(parts) >= 4 && zipcode == "" {
			// State is likely in parts[2], zip in parts[3]
			normalizedState = NormalizeState(parts[2])
			zipcode = parts[3]
		} else {
			// Try to parse state and zip from last part (e.g., "Portland, OR 97203")
			lastPart := parts[len(parts)-1]
			stateZipParts := strings.Fields(lastPart)
			if len(stateZipParts) >= 1 {
				normalizedState = NormalizeState(stateZipParts[0])
			}
			if len(stateZipParts) >= 2 {
				zipcode = stateZipParts[1]
			}
		}
	}

	// Normalize the components
	line1 = NormalizeAddressLine(line1, longHand)
	line1 = PostCleanAddrStr(line1)

	if line2 != "" {
		line2 = NormalizeOccupancyInLine(line2)
		line2 = PostCleanAddrStr(line2)
	}

	city = PostCleanAddrStr(city)
	normalizedState = NormalizeState(normalizedState)

	// Validate and format postal code
	var err error
	if zipcode != "" {
		zipcode, err = ValidateUSPostalCodeFormat(zipcode, addrStr)
		if err != nil {
			return nil, err
		}
	}

	// Validate parentheses are gone
	_, err = ValidateParensGroupsParsed(line1)
	if err != nil {
		return nil, err
	}

	return &Address{
		AddressLine1: line1,
		AddressLine2: line2,
		City:         city,
		State:        normalizedState,
		PostalCode:   zipcode,
	}, nil
}

// NormalizeAddrDict normalizes an address from a map.
func NormalizeAddrDict(addrDict map[string]string, strict bool, longHand bool) (*Address, error) {
	// Validate address components
	validated, err := ValidateAddressComponents(addrDict, strict)
	if err != nil {
		return nil, err
	}

	// Combine line 1 and line 2 for consistent processing
	addrStr := GetAddrLineStr(validated, true)

	postalCode := validated["postal_code"]
	city := validated["city"]
	state := validated["state"]

	var zipcode string
	if postalCode != "" {
		zipcode, err = ValidateUSPostalCodeFormat(postalCode, addrDict)
		if err != nil {
			return nil, err
		}
	}

	return NormalizeAddrStr(addrStr, "", city, state, zipcode, longHand)
}

// NormalizeAddressLine normalizes an address line by applying directional and street type normalizations.
func NormalizeAddressLine(line string, longHand bool) string {
	if line == "" {
		return ""
	}

	line = NormalizeDirectionalsInLine(line, longHand)
	line = NormalizeStreetTypesInLine(line, longHand)
	line = NormalizeNumberedStreets(line)

	return line
}

// NormalizeDirectionalsInLine normalizes directional words in an address line.
func NormalizeDirectionalsInLine(line string, longHand bool) string {
	words := strings.Fields(line)

	for i, word := range words {
		upperWord := strings.ToUpper(word)
		if abbr, exists := DirectionalReplacements[upperWord]; exists {
			if longHand {
				words[i] = upperWord // Keep it long form
			} else {
				words[i] = abbr
			}
		} else if longHand {
			// Check if it's already an abbreviation
			if longForm, exists := LonghandDirectionals[upperWord]; exists {
				words[i] = longForm
			}
		}
	}

	return strings.Join(words, " ")
}

// NormalizeStreetTypesInLine normalizes street type words in an address line.
func NormalizeStreetTypesInLine(line string, longHand bool) string {
	words := strings.Fields(line)

	for i, word := range words {
		upperWord := strings.ToUpper(word)
		if abbr, exists := StreetTypeAbbreviations[upperWord]; exists {
			if longHand {
				if longForm, ok := LonghandStreetTypes[abbr]; ok {
					words[i] = longForm
				} else {
					words[i] = abbr
				}
			} else {
				words[i] = abbr
			}
		} else if longHand {
			// Word might already be abbreviated
			if longForm, exists := LonghandStreetTypes[upperWord]; exists {
				words[i] = longForm
			}
		}
	}

	return strings.Join(words, " ")
}

// NormalizeNumberedStreets adds ordinal indicators to numbered streets.
func NormalizeNumberedStreets(line string) string {
	// Pattern to find numbered streets (e.g., "123 ST")
	pattern := regexp.MustCompile(`\b(\d+)\s+(ST|STREET|AVE|AVENUE|RD|ROAD)\b`)

	return pattern.ReplaceAllStringFunc(line, func(match string) string {
		parts := strings.Fields(match)
		if len(parts) >= 2 {
			num, err := strconv.Atoi(parts[0])
			if err == nil {
				ordinal := GetOrdinalIndicator(num)
				parts[0] = fmt.Sprintf("%d%s", num, ordinal)
				return strings.Join(parts, " ")
			}
		}
		return match
	})
}

// NormalizeOccupancyInLine normalizes occupancy types in a line.
func NormalizeOccupancyInLine(line string) string {
	if line == "" {
		return ""
	}

	words := strings.Fields(line)
	for i, word := range words {
		upperWord := strings.ToUpper(word)
		if abbr, exists := OccupancyTypeAbbreviations[upperWord]; exists {
			words[i] = abbr
		}
	}

	return strings.Join(words, " ")
}

// NormalizeState changes state string to accepted abbreviated format.
func NormalizeState(state string) string {
	if state == "" {
		return ""
	}

	upperState := strings.ToUpper(state)
	if abbr, exists := StateAbbreviations[upperState]; exists {
		return abbr
	}

	// If it's already a 2-letter code, return it
	if len(upperState) == 2 {
		return upperState
	}

	return state
}

// GetOrdinalIndicator gets the ordinal indicator suffix for a number.
//
// Ordinal numbers are words representing position or rank in a sequential
// order (1st, 2nd, 3rd, etc).
func GetOrdinalIndicator(number int) string {
	strNum := strconv.Itoa(number)
	lastDigit := strNum[len(strNum)-1:]

	// Special cases for 11, 12, 13
	if len(strNum) >= 2 {
		lastTwo := strNum[len(strNum)-2:]
		if lastTwo == "11" || lastTwo == "12" || lastTwo == "13" {
			return "TH"
		}
	}

	switch lastDigit {
	case "1":
		return "ST"
	case "2":
		return "ND"
	case "3":
		return "RD"
	default:
		return "TH"
	}
}

// GetAddrLineStr gets address 'line' elements as a single string.
func GetAddrLineStr(addrDict map[string]string, commaSeparate bool) string {
	var parts []string

	if addrDict["address_line_1"] != "" {
		parts = append(parts, addrDict["address_line_1"])
	}
	if addrDict["address_line_2"] != "" {
		parts = append(parts, addrDict["address_line_2"])
	}

	separator := " "
	if commaSeparate {
		separator = ", "
	}

	return strings.Join(parts, separator)
}

// ToMap converts an Address to a map[string]string
func (a *Address) ToMap() map[string]string {
	return map[string]string{
		"address_line_1": a.AddressLine1,
		"address_line_2": a.AddressLine2,
		"city":           a.City,
		"state":          a.State,
		"postal_code":    a.PostalCode,
	}
}
