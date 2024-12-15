package timer

import (
	"context"
	"time"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgeasybot/internal/orm"
)

type Config struct {
	Period time.Duration `yaml:"period"`
}

type TimerPoller struct {
	db      *orm.DB
	stopper chan struct{}
}

func NewTimerPoller(ctx context.Context, cfg Config, db *orm.DB) *TimerPoller {
	poller := TimerPoller{
		db:      db,
		stopper: make(chan struct{}),
	}
	go func() {
		ticker := time.NewTicker(cfg.Period)
		for {
			select {
			case <-poller.stopper:
				ticker.Stop()
				return
			case <-ticker.C:
				if num, err := db.ProcessTimers(ctx); err != nil {
					logging.S(ctx).Errorf("process timers: %s", err.Error())
				} else {
					logging.S(ctx).Debugf("updated timers: %d", num)
				}
			}
		}
	}()
	return &poller
}

func (t *TimerPoller) Shutdown() {
	t.stopper <- struct{}{}
}
