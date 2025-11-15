package scourgify

import (
	"testing"
)

func TestNormalizeState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Oregon", "OR"},
		{"or", "OR"},
		{"OREGON", "OR"},
		{"California", "CA"},
		{"calif", "CA"},
		{"OR", "OR"},
		{"", ""},
	}

	for _, test := range tests {
		result := NormalizeState(test.input)
		if result != test.expected {
			t.Errorf("NormalizeState(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestGetOrdinalIndicator(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "ST"},
		{2, "ND"},
		{3, "RD"},
		{4, "TH"},
		{11, "TH"},
		{12, "TH"},
		{13, "TH"},
		{21, "ST"},
		{22, "ND"},
		{23, "RD"},
		{101, "ST"},
		{113, "TH"},
	}

	for _, test := range tests {
		result := GetOrdinalIndicator(test.input)
		if result != test.expected {
			t.Errorf("GetOrdinalIndicator(%d) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestValidateUSPostalCodeFormat(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldError bool
	}{
		{"97203", "97203", false},
		{"2129", "02129", false},
		{"02129-44", "02129-0044", false},
		{"021290044", "02129-0044", false},
		{"97203-1234", "97203-1234", false},
		{"invalid", "", true},
		{"12345678901", "", true},
		{"02129-", "", true},
	}

	for _, test := range tests {
		result, err := ValidateUSPostalCodeFormat(test.input, nil)
		if test.shouldError {
			if err == nil {
				t.Errorf("ValidateUSPostalCodeFormat(%q) should have returned an error", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ValidateUSPostalCodeFormat(%q) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ValidateUSPostalCodeFormat(%q) = %q; want %q", test.input, result, test.expected)
			}
		}
	}
}

func TestNormalizeDirectionalsInLine(t *testing.T) {
	tests := []struct {
		input     string
		longHand  bool
		expected  string
	}{
		{"123 NORTH Main St", false, "123 N Main St"},
		{"123 N Main St", true, "123 NORTH Main St"},
		{"123 SOUTHWEST Main St", false, "123 SW Main St"},
		{"123 SW Main St", true, "123 SOUTHWEST Main St"},
	}

	for _, test := range tests {
		result := NormalizeDirectionalsInLine(test.input, test.longHand)
		if result != test.expected {
			t.Errorf("NormalizeDirectionalsInLine(%q, %v) = %q; want %q",
				test.input, test.longHand, result, test.expected)
		}
	}
}

func TestNormalizeStreetTypesInLine(t *testing.T) {
	tests := []struct {
		input     string
		longHand  bool
		expected  string
	}{
		{"123 Main Street", false, "123 Main ST"},
		{"123 Main ST", true, "123 Main STREET"},
		{"123 Main Avenue", false, "123 Main AVE"},
		{"123 Main BLVD", true, "123 Main BOULEVARD"},
	}

	for _, test := range tests {
		result := NormalizeStreetTypesInLine(test.input, test.longHand)
		if result != test.expected {
			t.Errorf("NormalizeStreetTypesInLine(%q, %v) = %q; want %q",
				test.input, test.longHand, result, test.expected)
		}
	}
}

func TestNormalizeAddrStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		longHand bool
		wantLine1 string
		wantCity  string
		wantState string
	}{
		{
			name:      "Basic address",
			input:     "123 Main Street, Portland, OR, 97203",
			longHand:  false,
			wantLine1: "123 MAIN ST",
			wantCity:  "PORTLAND",
			wantState: "OR",
		},
		{
			name:      "With directional",
			input:     "123 Southwest Main Street, Portland, OR",
			longHand:  false,
			wantLine1: "123 SW MAIN ST",
			wantCity:  "PORTLAND",
			wantState: "OR",
		},
		{
			name:      "Long hand format",
			input:     "123 SW Main St, Portland, OR",
			longHand:  true,
			wantLine1: "123 SOUTHWEST MAIN STREET",
			wantCity:  "PORTLAND",
			wantState: "OR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NormalizeAddrStr(tt.input, "", "", "", "", tt.longHand)
			if err != nil {
				t.Errorf("NormalizeAddrStr() error = %v", err)
				return
			}
			if addr.AddressLine1 != tt.wantLine1 {
				t.Errorf("AddressLine1 = %q; want %q", addr.AddressLine1, tt.wantLine1)
			}
			if addr.City != tt.wantCity {
				t.Errorf("City = %q; want %q", addr.City, tt.wantCity)
			}
			if addr.State != tt.wantState {
				t.Errorf("State = %q; want %q", addr.State, tt.wantState)
			}
		})
	}
}

func TestNormalizeAddrDict(t *testing.T) {
	input := map[string]string{
		"address_line_1": "123 southwest Main street",
		"address_line_2": "unit 2",
		"city":           "Boring",
		"state":          "or",
		"postal_code":    "97203",
	}

	addr, err := NormalizeAddrDict(input, true, false)
	if err != nil {
		t.Fatalf("NormalizeAddrDict() error = %v", err)
	}

	if addr.State != "OR" {
		t.Errorf("State = %q; want %q", addr.State, "OR")
	}
	if addr.City != "BORING" {
		t.Errorf("City = %q; want %q", addr.City, "BORING")
	}
	if addr.PostalCode != "97203" {
		t.Errorf("PostalCode = %q; want %q", addr.PostalCode, "97203")
	}
}

func TestValidateAddressComponents(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]string
		strict      bool
		shouldError bool
	}{
		{
			name: "Complete address strict",
			input: map[string]string{
				"address_line_1": "123 Main St",
				"city":           "Portland",
				"state":          "OR",
				"postal_code":    "97203",
			},
			strict:      true,
			shouldError: false,
		},
		{
			name: "Missing city strict",
			input: map[string]string{
				"address_line_1": "123 Main St",
				"state":          "OR",
				"postal_code":    "97203",
			},
			strict:      true,
			shouldError: true,
		},
		{
			name: "Only postal code non-strict",
			input: map[string]string{
				"address_line_1": "123 Main St",
				"postal_code":    "97203",
			},
			strict:      false,
			shouldError: false,
		},
		{
			name: "Missing line 1",
			input: map[string]string{
				"city":        "Portland",
				"state":       "OR",
				"postal_code": "97203",
			},
			strict:      true,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAddressComponents(tt.input, tt.strict)
			if tt.shouldError && err == nil {
				t.Error("ValidateAddressComponents() expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("ValidateAddressComponents() unexpected error: %v", err)
			}
		})
	}
}

func TestCleanUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  extra   spaces  ", "EXTRA SPACES"},
		{"Main Street", "MAIN STREET"},
		{"test@$%", "TEST"},             // Special chars removed
		{"test#123", "TEST#123"},        // # is in AllowedChars
		{"123 Main-St", "123 MAIN-ST"}, // Hyphen should be preserved in exclude list
	}

	for _, test := range tests {
		result := CleanUpper(test.input, AllowedChars, false)
		if result != test.expected {
			t.Errorf("CleanUpper(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}
