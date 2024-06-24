package statemachine

import (
	"context"
	"main/model"
)

var (
	EventPendingProcess Event = "PendingProcess"
)

// Calling try
func PendingProgressAction(ctx context.Context, trx *model.Transaction) (Event, error) {
	return "", nil
}
