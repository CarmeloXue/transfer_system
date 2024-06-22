package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmount(t *testing.T) {
	// value with no decimal
	str := "123"
	v, err := ParseFloat64String(str)
	assert.NoError(t, err)
	assert.Equal(t, float64(123), v)

	// value with decimal
	str = "123.123"
	v, err = ParseFloat64String(str)
	assert.NoError(t, err)
	assert.Equal(t, float64(123.123), v)

	// empty string
	str = ""
	v, err = ParseFloat64String(str)
	assert.EqualError(t, ErrAmountInvalidFormat, err.Error())
	assert.Equal(t, float64(0), v)

	// invalid format
	str = "123,123"
	v, err = ParseFloat64String(str)
	assert.EqualError(t, ErrAmountInvalidFormat, err.Error())
	assert.Equal(t, float64(0), v)

	// negative
	str = "-10"
	v, err = ParseFloat64String(str)
	assert.Error(t, err)
	assert.EqualError(t, ErrAmountNegative, err.Error())
	assert.Equal(t, float64(0), v)
}
