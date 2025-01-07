package sm

import (
	"context"
	"sync"
	"time"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"github.com/baldisbk/tgeasybot/internal/orm"
)

type PostponeEvent struct{}
type DropEvent struct{}

type Worker struct {
	db     *orm.DB
	client tgapi.TGClient

	stopper chan struct{}
}

type WorkerPool []*Worker

func NewWorkerPool(ctx context.Context, cfg Config, client tgapi.TGClient, db *orm.DB) WorkerPool {
	var res WorkerPool
	for i := 0; i < cfg.Workers; i++ {
		worker := &Worker{
			db:      db,
			client:  client,
			stopper: make(chan struct{}),
		}
		go func() {
			ticker := time.NewTicker(cfg.Period)
			for {
				select {
				case <-worker.stopper:
					ticker.Stop()
					return
				case <-ticker.C:
					logging.S(ctx).Debugf("Fetching events...")
					events, user, err := worker.db.FetchEvents(ctx)
					if err != nil {
						logging.S(ctx).Errorf("error fetching events: %s", err.Error())
						continue
					}
					if len(events) == 0 {
						continue
					}
					if user == nil {
						logging.S(ctx).Errorf("user not found")
						continue
					}
					logging.S(ctx).Debugf("Fetched %d events for user %s", len(events), user.Name)
					doer, err := NewDoer(user, worker.db, worker.client)
					if err != nil {
						logging.S(ctx).Errorf("error creating user: %s", err.Error())
						continue
					}
					for _, event := range events {
						logging.S(ctx).Debugf("Processing event %q...", event.Payload)
						in, err := UnmarshalEvent(event)
						if err != nil {
							logging.S(ctx).Errorf("error unmarshaling event: %s", err.Error())
							continue
						}
						sm := MakeStateMachine(doer)
						out, err := sm.Run(ctx, in)
						if err != nil {
							logging.S(ctx).Errorf("error fetching events: %s", err.Error())
							continue
						}
						switch e := out.(type) {
						case PostponeEvent:
							logging.S(ctx).Debugf("Event postponed")
							// change nothing, process event another time
							continue
						case DropEvent:
							logging.S(ctx).Debugf("Event discarded")
							// change nothing but delete event
							event.Busy = -1
							continue
						case string:
							user.Contents = e
						default:
							// unexpected - probably unprocessed input - ignore
							logging.S(ctx).Warnf("Unknown event: %#v", out)
							continue
						}
						event.Busy = -1 // drop event, it is processed
						user.State = sm.State()
						logging.S(ctx).Debugf("Event processed, new state %q, contents %q", user.State, user.Contents)
					}
					if err := worker.db.ProcessEvent(ctx, user, events); err != nil {
						logging.S(ctx).Errorf("error processing event: %s", err.Error())
					}
				}
			}
		}()
		res = append(res, worker)
	}
	return res
}

func (p *WorkerPool) Shutdown() {
	var wg sync.WaitGroup
	wg.Add(len(*p))
	for _, worker := range *p {
		go func() {
			defer wg.Done()
			close(worker.stopper)
		}()
	}
	wg.Wait()
}
