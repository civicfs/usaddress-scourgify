# usaddress-scourgify

A Go library for cleaning/normalizing US addresses following USPS pub 28 and RESO guidelines.

**Note:** This is a Go port of the original Python library. This version provides core normalization functionality but uses simplified address parsing compared to the Python version which relies on the `usaddress` library's probabilistic CRF-based parser.

## Features

- Normalize US addresses to USPS Publication 28 standards
- Standardize directional indicators (e.g., "Southwest" → "SW")
- Abbreviate street types (e.g., "Street" → "ST")
- Normalize state names to 2-character abbreviations
- Format postal codes to ZIP or ZIP+4 format
- Add ordinal indicators to numbered streets (e.g., "113" → "113TH")
- Support for both abbreviated and long-hand output formats
- Clean special characters and handle edge cases

## Installation

```bash
go get github.com/GreenBuildingRegistry/usaddress-scourgify
```

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/GreenBuildingRegistry/usaddress-scourgify"
)

func main() {
	// Normalize an address string
	addr, err := scourgify.NormalizeAddrStr(
		"123 southwest Main street, Boring, or, 97203",
		"",     // line2
		"",     // city (will be parsed from string)
		"",     // state (will be parsed from string)
		"",     // zipcode (will be parsed from string)
		false,  // longHand
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Address Line 1: %s\n", addr.AddressLine1)
	fmt.Printf("City: %s\n", addr.City)
	fmt.Printf("State: %s\n", addr.State)
	fmt.Printf("Postal Code: %s\n", addr.PostalCode)
}
```

### Normalize from Map

```go
addrMap := map[string]string{
	"address_line_1": "123 southwest Main street",
	"address_line_2": "unit 2",
	"city":           "Boring",
	"state":          "or",
	"postal_code":    "97203",
}

addr, err := scourgify.NormalizeAddrDict(addrMap, true, false)
if err != nil {
	log.Fatal(err)
}

// Output:
// {
//     "address_line_1": "123 SW MAIN ST",
//     "address_line_2": "UNIT 2",
//     "city": "BORING",
//     "state": "OR",
//     "postal_code": "97203"
// }
```

### Long-Hand Format

```go
addr, err := scourgify.NormalizeAddrStr(
	"123 SW Main St, Boring, OR, 97203",
	"",
	"",
	"",
	"",
	true, // longHand = true
)

// Output:
// {
//     "address_line_1": "123 SOUTHWEST MAIN STREET",
//     "city": "BORING",
//     "state": "OR",
//     "postal_code": "97203"
// }
```

### Using NormalizeAddressRecord

```go
// Works with both strings and maps
addr1, err := scourgify.NormalizeAddressRecord(
	"123 SW Main St, Boring, OR, 97203",
	false, // strict
	false, // longHand
)

addr2, err := scourgify.NormalizeAddressRecord(
	map[string]string{
		"address_line_1": "123 SW Main St",
		"city":           "Boring",
		"state":          "OR",
		"postal_code":    "97203",
	},
	true,  // strict
	false, // longHand
)
```

## API Reference

### Main Functions

#### `NormalizeAddressRecord(address interface{}, strict bool, longHand bool) (*Address, error)`

Normalizes an address from either a string or map.

- `address`: Either a string or `map[string]string` containing address data
- `strict`: If true, requires city, state, AND postal code. If false, requires either postal code OR (city and state)
- `longHand`: If true, outputs full directional and street type names instead of abbreviations

Returns a normalized `Address` struct or an error.

#### `NormalizeAddrStr(addrStr, line2, city, state, zipcode string, longHand bool) (*Address, error)`

Normalizes an address string with optional pre-parsed components.

#### `NormalizeAddrDict(addrDict map[string]string, strict bool, longHand bool) (*Address, error)`

Normalizes an address from a map.

### Address Struct

```go
type Address struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
}
```

### Error Types

- `IncompleteAddressError`: Address is missing required elements
- `AmbiguousAddressError`: Address contains ambiguous elements
- `UnParseableAddressError`: Unable to parse address
- `AddressValidationError`: Address format validation failed

## Postal Code Normalization

Postal codes are normalized to US ZIP or ZIP+4 format with zero-padding:

- `2129` → `02129`
- `02129-44` → `02129-0044`
- `021290044` → `02129-0044`

Invalid postal codes (wrong length, invalid characters, etc.) will raise an `AddressValidationError`.

## Differences from Python Version

This Go implementation provides the same core normalization features as the Python version but with some differences:

1. **Address Parsing**: The Python version uses the `usaddress` library which employs conditional random fields (CRF) for probabilistic address parsing. This Go version uses simplified regex and rule-based parsing.

2. **Geocoder Support**: The Python version includes `get_geocoder_normalized_addr()` which uses Google's Geocoder API. This feature is not included in the Go version.

3. **Custom Configuration**: The Python version supports YAML configuration files for customizing constants. This Go version uses the default constants defined in code.

4. **Class Interface**: The Python version provides a `NormalizeAddress` class. This Go version uses functions with a simple `Address` struct.

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build
```

## Contributing

Create a new branch to hold your changes. Include a comment explaining the issue your pull request solves. Make sure all tests pass.

## License

usaddress-scourgify is released under the terms of the MIT license. Full details in LICENSE file.

## Changelog

### Version 1.0.0 (Go Port)

- Initial Go implementation
- Core address normalization functionality
- Support for abbreviated and long-hand formats
- Postal code validation and formatting
- State abbreviation normalization
- Street type and directional normalization
- Ordinal indicator support for numbered streets

## Credits

Originally developed for the greenbuildingregistry project.

- Original Python version by Fable Turas
- Maintained by GreenBuildingRegistry
- Go port by GreenBuildingRegistry

## References

- [USPS Publication 28](https://pe.usps.com/text/pub28/welcome.htm) - Postal Addressing Standards
- [RESO Data Dictionary](https://www.reso.org/data-dictionary/) - Real Estate Standards Organization
