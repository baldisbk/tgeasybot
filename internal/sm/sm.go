package sm

import (
	"context"

	"github.com/baldisbk/tgbot/pkg/statemachine"
)

const (
	noState    = ""
	startState = "start"
)

type PostponeEvent struct{}
type DropEvent struct{}

func CompositePredicate(preds ...statemachine.SMPredicate) statemachine.SMPredicate {
	return func(ctx context.Context, s string, i interface{}) bool {
		for _, p := range preds {
			if !p(ctx, s, i) {
				return false
			}
		}
		return true
	}
}

func IsMessage(ctx context.Context, _ string, input interface{}) bool {
	if input == nil {
		return false
	}
	_, ok := input.(*Message)
	return ok
}

func IsCallback(ctx context.Context, _ string, input interface{}) bool {
	if input == nil {
		return false
	}
	_, ok := input.(*Message)
	return ok
}

func MessageContains(origin string) statemachine.SMPredicate {
	return func(ctx context.Context, _ string, input interface{}) bool {
		return input.(*Message).Text == origin
	}
}

func IsTimer(id int64) statemachine.SMPredicate {
	return func(ctx context.Context, _ string, input interface{}) bool {
		if input == nil {
			return false
		}
		timer, ok := input.(*Timer)
		return ok && timer.ID == id
	}
}

func MakeStateMachine(doer *Doer) statemachine.Machine {
	state := doer.User.State
	if state == "" {
		state = noState
	}
	return statemachine.NewSM(state, []statemachine.Transition{
		{
			Source:      noState,
			Destination: startState,
			Predicate:   CompositePredicate(IsMessage, MessageContains("/start")),
			Callback:    doer.Greeting,
		},
		{
			Source:      startState,
			Destination: startState,
			Predicate:   IsTimer(greetingTimer),
			Callback:    doer.TimedGreeting,
		},
	}, true)
}
