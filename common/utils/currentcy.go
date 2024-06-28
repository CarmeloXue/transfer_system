package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrOverflow      = errors.New("amount overflow")
	ErrTooManyDigits = errors.New("too many digits. only 6 digit allowed")
	ErrNegativeValue = errors.New("negative value")
)

// ParseString parses a string into an int64, allowing up to 6 digits after the decimal point.
func ParseString(s string) (int64, error) {
	s = strings.Trim(s, " ")
	// Check if the string has a decimal point
	parts := strings.Split(s, ".")
	if len(parts) == 2 {
		if len(parts[1]) > 6 {
			return 0, ErrTooManyDigits
		}
	}

	// Parse the float value
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		if strings.Contains(err.Error(), "invalid syntax") {
			return 0, err
		}
		return 0, ErrOverflow
	}

	// Scale the float to an int64
	result := int64(f * 1e6)
	return result, nil
}

// ParseFloat parses a float64 into an int64, allowing up to 6 digits after the decimal point.
func ParseFloat(f float64) (int64, error) {
	s := fmt.Sprintf("%.6f", f)

	parsedValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if parsedValue != f {
		return 0, ErrTooManyDigits
	}
	parts := strings.Split(s, ".")
	if len(parts) == 2 {
		if len(parts[1]) > 6 {
			return 0, ErrTooManyDigits
		}
	}

	// Scale the float to an int64
	result := int64(f * 1e6)
	return result, nil
}

// FormatInt formats an int64 to a float string with 6 digits after the decimal point.
func FormatInt(i int64) string {
	return fmt.Sprintf("%.6f", float64(i)/1e6)
}

// SafeAdd safely adds multiple int64 values, checking for overflow.
func SafeAdd(nums ...int64) (int64, error) {
	var sum int64
	for _, num := range nums {
		// Check for overflow
		if (num > 0 && sum > math.MaxInt64-num) || (num < 0 && sum < math.MinInt64-num) {
			return 0, ErrOverflow
		}
		sum += num
	}
	if sum < 0 {
		return 0, ErrNegativeValue
	}
	return sum, nil
}
