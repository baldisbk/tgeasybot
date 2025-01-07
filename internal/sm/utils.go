package sm

import (
	"context"
	"encoding/json"
	"time"

	"github.com/baldisbk/tgbot/pkg/statemachine"
	"github.com/baldisbk/tgeasybot/internal/orm"
	"golang.org/x/xerrors"
)

// ======== tools ========

func (d *Doer) sendMessage(ctx context.Context, msg string) error {
	_, err := d.api.SendMessage(ctx, uint64(d.User.ID), msg)
	return err
}

func (d *Doer) setTimer(ctx context.Context, id int64, t time.Time, repeat bool) error {
	return d.db.SetupTimer(ctx, &orm.Timer{
		ID:         id,
		TS:         t.Unix(),
		UserID:     d.User.ID,
		Repeatable: repeat,
	})
}

func (d *Doer) ackCallback(ctx context.Context, cb *CallbackQuery) error {
	return d.api.AnswerCallback(ctx, cb.Id)
}

func (d *Doer) dbState() (interface{}, error) {
	b, err := json.Marshal(d.state)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// ======== wrappers ========

func (d *Doer) Msg(ff ...func(context.Context, *Message) error) statemachine.SMCallback {
	return func(ctx context.Context, i interface{}) (interface{}, error) {
		msg, ok := i.(*Message)
		if !ok {
			return nil, xerrors.Errorf("not a message: %v", i)
		}
		for _, f := range ff {
			if err := f(ctx, msg); err != nil {
				return nil, err
			}
		}
		return d.dbState()
	}
}

func (d *Doer) Cb(ff ...func(context.Context, *CallbackQuery) error) statemachine.SMCallback {
	return func(ctx context.Context, i interface{}) (interface{}, error) {
		cb, ok := i.(*CallbackQuery)
		if !ok {
			return nil, xerrors.Errorf("not a callback: %v", i)
		}
		if err := d.ackCallback(ctx, cb); err != nil {
			return nil, xerrors.Errorf("ack callback: %w", err)
		}
		for _, f := range ff {
			if err := f(ctx, cb); err != nil {
				return nil, err
			}
		}
		return d.dbState()
	}
}

func (d *Doer) T(ff ...func(context.Context, *Timer) error) statemachine.SMCallback {
	return func(ctx context.Context, i interface{}) (interface{}, error) {
		t, ok := i.(*Timer)
		if !ok {
			return nil, xerrors.Errorf("not a timer: %v", i)
		}
		for _, f := range ff {
			if err := f(ctx, t); err != nil {
				return nil, err
			}
		}
		return d.dbState()
	}
}
