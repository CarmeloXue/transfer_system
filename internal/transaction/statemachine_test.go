package transaction

import (
	"context"
	"errors"
	"main/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	EventStart  Event = "start"
	EventMiddle Event = "middle"
	EventEnd    Event = "end"
)

func Test_StateMachine(t *testing.T) {

	var sm = StateMachine{
		EventStart: State{
			Action:         StartAction,
			Name:           "start",
			TimeoutSeconds: 1,
			MaxCallTimes:   1,
			IsFinal:        func() bool { return false },
		},
		EventMiddle: State{
			Name:           "middle",
			Action:         MiddleAction,
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal:        func() bool { return false },
		},
		EventEnd: State{
			Name:           "end",
			Action:         EndAction,
			TimeoutSeconds: 1,
			MaxCallTimes:   1,
			IsFinal:        func() bool { return true },
		},
	}
	tx := model.Transaction{
		TransactionID: "123",
	}
	err := sm.Start(context.Background(), EventStart, &tx)
	assert.EqualError(t, errors.New("middle error"), err.Error())
	assert.Equal(t, 3, tx.Retries)
}

func StartAction(ctx context.Context, txn *model.Transaction) (Event, error) {
	return EventMiddle, nil
}
func MiddleAction(ctx context.Context, txn *model.Transaction) (Event, error) {
	txn.Retries++
	select {
	case <-ctx.Done():
		return "", errors.New("middle error")
	case <-time.After(2 * time.Second):
		return "", nil
	}
}
func EndAction(ctx context.Context, txn *model.Transaction) (Event, error) {
	time.Sleep(3)
	return "", errors.New("End error")
}
