package utils

import (
	"errors"
	"math"
	"strconv"
)

var (
	ErrAmountNegative        = errors.New("invalid amount. cannot be negative")
	ErrAmountInvalidFormat   = errors.New("invalid amount. invalid format")
	ErrAmountExceedsMaxValue = errors.New("invalid amount. exceeds max value")
)

// ParseFloat64String validates if a string can be parsed into a float64.
func ParseFloat64String(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, ErrAmountInvalidFormat
	}
	if v < 0 {
		return 0, ErrAmountNegative
	}

	if v > math.MaxFloat64 || v < -math.MaxFloat64 {
		return 0, ErrAmountExceedsMaxValue
	}

	return v, nil
}

// StringToFloat64 converts a string to float64.
// It assumes the string is already validated or trusted.
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
