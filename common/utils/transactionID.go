package utils

import (
	"github.com/google/uuid"
)

func GenerateTransactionID() string {
	return uuid.New().String()
}
