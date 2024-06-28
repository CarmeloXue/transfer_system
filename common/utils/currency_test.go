package utils

import (
	"fmt"
	"math"
	"testing"
)

func TestSafeAdd(t *testing.T) {
	tests := []struct {
		a, b   int64
		result int64
		err    error
	}{
		{1, 2, 3, nil},
		{math.MaxInt64, 1, 0, ErrOverflow},
		{math.MinInt64, -1, 0, ErrOverflow},
		{math.MaxInt64, -1, math.MaxInt64 - 1, nil},
		{math.MinInt64, 1, math.MinInt64 + 1, nil},
	}

	for _, test := range tests {
		res, err := SafeAdd(test.a, test.b)
		if res != test.result || err != test.err {
			t.Errorf("SafeAdd(%d, %d) = (%d, %v), want (%d, %v)", test.a, test.b, res, err, test.result, test.err)
		}
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		input  string
		result int64
		err    error
	}{
		{"1.123456 ", 1123456, nil},
		{"123456", 123456000000, nil},
		{" 123456", 123456000000, nil},
		{" 123456    ", 123456000000, nil},
		{" 1.1234567 ", 0, ErrTooManyDigits},
		{"abc", 0, fmt.Errorf("strconv.ParseFloat: parsing \"abc\": invalid syntax")},
		{"201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999920123999999201239999992012399999", 0, ErrOverflow},
	}

	for _, test := range tests {
		res, err := ParseString(test.input)
		if res != test.result || (err != nil && err.Error() != test.err.Error()) {
			t.Errorf("ParseString(%s) = (%d, %v), want (%d, %v)", test.input, res, err, test.result, test.err)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input  float64
		result int64
		err    error
	}{
		{1.123456, 1123456, nil},
		{123456.0, 123456000000, nil},
		{1.1234567, 0, ErrTooManyDigits},
		{123450.00, 123450000000, nil},
	}

	for _, test := range tests {
		res, err := ParseFloat(test.input)
		if res != test.result || err != test.err {
			t.Errorf("ParseFloat(%f) = (%d, %v), want (%d, %v)", test.input, res, err, test.result, test.err)
		}
	}
}
