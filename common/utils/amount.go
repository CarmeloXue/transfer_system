package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrAmountNegative        = errors.New("invalid amount. cannot be negative")
	ErrAmountInvalidFormat   = errors.New("invalid amount. invalid format")
	ErrAmountExceedsMaxValue = errors.New("invalid amount. exceeds max value")
)

func ParseFloat64String(input string) (float64, error) {
	// Remove any leading or trailing whitespace
	input = strings.TrimSpace(input)

	// Split the input by the decimal point
	parts := strings.Split(input, ".")
	if len(parts) > 2 {
		return 0, ErrAmountInvalidFormat
	}

	// Parse the integer part
	intPart := parts[0]
	intValue, err := strconv.Atoi(intPart)
	if err != nil {
		return 0, ErrAmountInvalidFormat
	}
	if intValue < 0 {
		return 0, ErrAmountNegative
	}

	// Convert the input string to a float
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse input as float: %v", err)
	}

	return value, nil
}

// StringToFloat64 converts a string to float64.
// It assumes the string is already validated or trusted.
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
