package sm

import (
	"context"

	"github.com/baldisbk/tgbot/pkg/statemachine"
)

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

func IsCallback(check string) statemachine.SMPredicate {
	return func(ctx context.Context, state string, input interface{}) bool {
		if input == nil {
			return false
		}
		cb, ok := input.(*CallbackQuery)
		return ok && (check == "*" || cb.Button == check)
	}
}

func IsNumButton(ctx context.Context, state string, input interface{}) bool {
	if input == nil {
		return false
	}
	cb, ok := input.(*CallbackQuery)
	return ok && buttonNum(cb.Button) != -1
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
