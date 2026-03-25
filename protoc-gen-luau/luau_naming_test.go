package main

import "testing"

func TestSnakeToCamel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"type_id", "typeId"},
		{"value", "value"},
		{"format_patterns", "formatPatterns"},
		{"value_sub_fields", "valueSubFields"},
		{"sub_field_keys_in_order", "subFieldKeysInOrder"},
		{"all_option_keys_in_order", "allOptionKeysInOrder"},
		{"iso_country_code", "isoCountryCode"},
		{"aba", "aba"},
		{"swift", "swift"},
	}
	for _, tt := range tests {
		got := snakeToCamel(tt.in)
		if got != tt.want {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
