package transaction

import (
	"context"
	"errors"
	"main/model"
	"time"
)

type Event string

var (
	EventPendingProcess Event = "PendingProcess"
	EventPendingFulfil  Event = "PendingFulfil"
	EventPendingRefund  Event = "PendingRefund"
	EventPendingFailed  Event = "PendingFailed"
	EventFulfiled       Event = "Fulfiled"
	EventRefund         Event = "Refund"
	EventFailed         Event = "Failed"
)

type State struct {
	Name           string
	Action         func(ctx context.Context, transaction *model.Transaction) (Event, error)
	MaxCallTimes   int
	TimeoutSeconds int
	IsFinal        func() bool
}

type StateMachine map[Event]State

func GetStateMachine() StateMachine {
	transactionSM := map[Event]State{
		EventPendingProcess: State{
			MaxCallTimes:   0,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return true
			},
		},
		EventPendingFulfil: State{
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return false
			},
		},
		EventPendingRefund: State{
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return false
			},
		},
		EventPendingFailed: State{
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return false
			},
		},
		EventFulfiled: State{
			MaxCallTimes:   1,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return true
			},
		},
		EventFailed: State{
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return true
			},
		},
		EventRefund: State{
			MaxCallTimes:   3,
			TimeoutSeconds: 1,
			IsFinal: func() bool {
				return true
			},
		},
	}
	return transactionSM
}

func (sm StateMachine) Start(ctx context.Context, startEvent Event, trx *model.Transaction) error {
	currentEvent := startEvent

	for {
		var err error
		currentState, ok := sm[currentEvent]
		if !ok {
			return errors.New("unknown event")
		}

		// Create a context with timeout based on currentState.TimeoutSeconds
		actionCtx, cancel := context.WithTimeout(ctx, time.Duration(currentState.TimeoutSeconds)*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		// Retry loop based on MaxRetries
		for retry := 0; retry < currentState.MaxCallTimes; retry++ {
			var nextEvent Event
			nextEvent, err = currentState.Action(actionCtx, trx)
			if retry < currentState.MaxCallTimes {
				continue
			}
			if err != nil {
				return err
			}
			currentEvent = nextEvent
		}

		if err != nil {
			return err
		}

		// Check if currentState is final
		if currentState.IsFinal() {
			return nil
		}
	}
}
