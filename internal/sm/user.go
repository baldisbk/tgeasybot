package sm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgeasybot/internal/orm"
	"golang.org/x/xerrors"
)

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

// ======== callbacks ========

func (d *Doer) Greeting(ctx context.Context, _ *Message) error {
	if err := d.sendMessage(ctx, fmt.Sprintf("Hello, %s! Let's go!", d.User.Name)); err != nil {
		return xerrors.Errorf("send: %w", err)
	}
	return nil
}

func (d *Doer) SetTimeout(ctx context.Context, _ interface{}) (interface{}, error) {
	if err := d.setTimer(ctx, timeoutTimer, time.Now().Add(3*time.Minute), false); err != nil {
		return nil, xerrors.Errorf("timer: %w", err)
	}
	return d.dbState()
}

func (d *Doer) MainMenu(ctx context.Context, _ interface{}) (interface{}, error) {
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID),
		fmt.Sprintf("Whatcha gonna do, %s?", d.User.Name),
		d.state.LastMessage, mainMenu)
	if err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID

	return d.dbState()
}

func (d *Doer) EditMenu(ctx context.Context, _ interface{}) (interface{}, error) {
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID),
		fmt.Sprintf("Setup your achievements, %s?", d.User.Name),
		d.state.LastMessage, editMenu)
	if err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID

	return d.dbState()
}

func (d *Doer) ChangeMenu(ctx context.Context, input interface{}) (interface{}, error) {
	return d.listMenu(ctx, "Choose achievement to change:")
}

func (d *Doer) DeleteMenu(ctx context.Context, input interface{}) (interface{}, error) {
	return d.listMenu(ctx, "Choose achievement to delete:")
}

func (d *Doer) CheckMenu(ctx context.Context, input interface{}) (interface{}, error) {
	return d.listMenu(ctx, "Choose achievement to check:")
}

func (d *Doer) StatMenu(ctx context.Context, input interface{}) (interface{}, error) {
	return d.listMenu(ctx, "Choose achievement to show:")
}

func (d *Doer) listMenu(ctx context.Context, msg string) (interface{}, error) {
	lines := []string{msg}
	for i := d.state.StatIndex; i < len(d.state.Stats) && i-d.state.StatIndex < len(numButtons); i++ {
		lines = append(lines, fmt.Sprintf("(%d) %s", i+1, d.state.Stats[i].Name))
	}
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID),
		strings.Join(lines, "\n"),
		d.state.LastMessage,
		listMenu(d.state.StatIndex, len(d.state.Stats)))
	if err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID
	return d.dbState()
}

func (d *Doer) Scroll(ctx context.Context, cb *CallbackQuery) error {
	switch cb.Button {
	case prevButton:
		d.state.StatIndex -= len(numButtons)
	case nextButton:
		d.state.StatIndex += len(numButtons)
	}
	return nil
}

func (d *Doer) Add(ctx context.Context, input interface{}) (interface{}, error) {
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID),
		"Input achievement name:",
		d.state.LastMessage, backMenu)
	if err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID

	return d.dbState()
}

func (d *Doer) confirm(ctx context.Context, msg string) (interface{}, error) {
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID), msg,
		d.state.LastMessage, yesNoMenu)
	if err != nil {
		return nil, xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID
	return d.dbState()
}

func (d *Doer) AddConfirm(ctx context.Context, msg *Message) error {
	d.state.InputStat = UserStat{Name: msg.Text}
	_, err := d.confirm(ctx, fmt.Sprintf("Adding achievement %q...\nConfirm?", d.state.InputStat.Name))
	return err
}

func (d *Doer) ChangeConfirm(ctx context.Context, msg *Message) error {
	d.state.InputStat.Name = msg.Text
	_, err := d.confirm(ctx, fmt.Sprintf("Adding achievement %q...\nConfirm?", d.state.InputStat.Name))
	return err
}

func (d *Doer) DelConfirm(ctx context.Context, cb *CallbackQuery) error {
	d.state.StatIndex += buttonNum(cb.Button)
	ach := d.state.Stats[d.state.StatIndex]
	_, err := d.confirm(ctx, fmt.Sprintf("Removing achievement %q...\nConfirm?", ach.Name))
	return err
}

func (d *Doer) Change(ctx context.Context, cb *CallbackQuery) error {
	d.state.StatIndex += buttonNum(cb.Button)
	d.state.InputStat = d.state.Stats[d.state.StatIndex]
	msgID, err := d.api.EditInputKeyboard(
		ctx, uint64(d.User.ID),
		fmt.Sprintf("Input achievement name (now: %q):", d.state.InputStat.Name),
		d.state.LastMessage, backMenu)
	if err != nil {
		return xerrors.Errorf("send: %w", err)
	}
	d.state.LastMessage = msgID
	return nil
}

func (d *Doer) DoAdd(ctx context.Context, input interface{}) (interface{}, error) {
	d.state.Stats = append(d.state.Stats, d.state.InputStat)
	d.state.UserTemporaryState = UserTemporaryState{}
	return d.dbState()
}

func (d *Doer) DoChange(ctx context.Context, input interface{}) (interface{}, error) {
	d.state.Stats[d.state.StatIndex] = d.state.InputStat
	d.state.UserTemporaryState = UserTemporaryState{}
	return d.dbState()
}

func (d *Doer) DoDel(ctx context.Context, input interface{}) (interface{}, error) {
	d.state.Stats = append(d.state.Stats[:d.state.StatIndex], d.state.Stats[d.state.StatIndex+1:]...)
	d.state.UserTemporaryState = UserTemporaryState{}
	return d.dbState()
}

func (d *Doer) DropKB(ctx context.Context, cb *CallbackQuery) error {
	d.api.DropKeyboard(ctx, uint64(d.User.ID), "The abort")
	return nil
}
