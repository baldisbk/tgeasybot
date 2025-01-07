package poller

import (
	"context"
	"encoding/json"
	"time"

	"golang.org/x/xerrors"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"

	"github.com/baldisbk/tgeasybot/internal/orm"
	"github.com/baldisbk/tgeasybot/internal/sm"
)

type Poller struct {
	Client tgapi.TGClient
	db     *orm.DB

	stopper chan struct{}
}

type Config struct {
	PollPeriod time.Duration `yaml:"period"`
}

func NewPoller(ctx context.Context, cfg Config, client tgapi.TGClient, db *orm.DB) *Poller {
	poller := &Poller{
		Client:  client,
		db:      db,
		stopper: make(chan struct{}),
	}
	go poller.run(ctx, cfg.PollPeriod)
	return poller
}

func (p *Poller) Shutdown() { close(p.stopper) }

func (p *Poller) do(ctx context.Context) (int, error) {
	state, err := p.db.GetState(ctx)
	if err != nil {
		return 0, xerrors.Errorf("get state: %w", err)
	}
	offset := uint64(state.Offset)
	if offset != 0 {
		offset++
	}
	upds, _, err := p.Client.GetUpdates(ctx, offset)
	if err != nil {
		return 0, xerrors.Errorf("get updates: %w", err)
	}
	var events []orm.Event
	for _, upd := range upds {
		var event orm.Event
		var contents []byte
		var err error
		switch {
		case upd.Message != nil:
			contents, _ = json.Marshal(sm.Message{
				From: sm.User{FirstName: upd.Message.From.FirstName},
				Text: upd.Message.Text,
				Date: int64(upd.Message.Date),
			})
			event = orm.Event{
				ID:       int64(upd.UpdateId),
				TS:       int64(upd.Message.Date),
				UserID:   int64(upd.Message.From.Id),
				Type:     orm.EventTypeMessage,
				Payload:  string(contents),
				UserName: upd.Message.From.FirstName,
			}
		case upd.CallbackQuery != nil:
			contents, err = json.Marshal(sm.CallbackQuery{
				From:   sm.User{FirstName: upd.CallbackQuery.From.FirstName},
				Button: upd.CallbackQuery.Data,
				Id:     upd.CallbackQuery.Id,
			})
			event = orm.Event{
				ID:       int64(upd.UpdateId),
				TS:       time.Now().Unix(),
				UserID:   int64(upd.CallbackQuery.From.Id),
				Type:     orm.EventTypeCallback,
				Payload:  string(contents),
				UserName: upd.CallbackQuery.From.FirstName,
			}
		}
		if err != nil {
			logging.S(ctx).Errorf("error marshalling payload: %s", err.Error())
			continue
		}
		events = append(events, event)
	}
	return p.db.RegisterEvents(ctx, events)
}

func (p *Poller) run(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			if num, err := p.do(ctx); err != nil {
				logging.S(ctx).Errorf("Error processing updates: %#v", err)
			} else if num != 0 {
				logging.S(ctx).Debugf("Updates processed: %d", num)
			}
		case <-p.stopper:
			return
		case <-ctx.Done():
			return
		}
	}
}
