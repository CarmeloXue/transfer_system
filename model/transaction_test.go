package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Display(t *testing.T) {
	tx := Transaction{
		Amount: 1000000,
	}

	(&tx).FormatForDisplay()

	assert.Equal(t, "1.000000", tx.TransactionAmount)

	bs, _ := json.Marshal(tx)
	assert.True(t, !strings.Contains(string(bs), "amount:"))

}
