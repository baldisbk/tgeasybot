package sm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgeasybot/internal/orm"
	"golang.org/x/xerrors"
)

const (
	greetingTimer int64 = iota
)

type UserState struct{}

func (s *UserState) Marshal() (string, error) {
	b, err := json.Marshal(*s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func StateUnmarshal(from string) (*UserState, error) {
	var s UserState
	if from == "" {
		return &s, nil
	}
	err := json.Unmarshal([]byte(from), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type Doer struct {
	User *orm.User

	state *UserState

	db  *orm.DB
	api tgapi.TGClient
}

func NewDoer(user *orm.User, db *orm.DB, api tgapi.TGClient) (*Doer, error) {
	state, err := StateUnmarshal(user.Contents)
	if err != nil {
		return nil, xerrors.Errorf("bad contents: %w", err)
	}
	return &Doer{
		User:  user,
		state: state,
		db:    db,
		api:   api,
	}, nil
}

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

func (d *Doer) Greeting(ctx context.Context, input interface{}) (interface{}, error) {
	msg := input.(*Message)
	if err := d.sendMessage(ctx, fmt.Sprintf("Hello, %s! You wrote: %q", d.User.Name, msg.Text)); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	if err := d.setTimer(ctx, greetingTimer, time.Now().Add(3*time.Minute), false); err != nil {
		return nil, xerrors.Errorf("timer: %w", err)
	}

	return nil, nil
}

func (d *Doer) TimedGreeting(ctx context.Context, input interface{}) (interface{}, error) {
	if err := d.sendMessage(ctx, fmt.Sprintf("Hello again, %s!", d.User.Name)); err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}

	return "{}", nil
}
