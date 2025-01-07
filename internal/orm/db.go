package orm

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/xerrors"
)

const defaultDeprecate = time.Minute

type DB struct {
	pool      *pgxpool.Pool
	deprecate time.Duration // events/users are not busy anymore
}

func NewDB(ctx context.Context, cfg Config) (*DB, error) {
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(cfg.CertFile)
	if err != nil {
		return nil, xerrors.Errorf("error loading %s: %w", cfg.CertFile, err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return nil, xerrors.Errorf("failed to append PEM")
	}

	connstring := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=verify-full target_session_attrs=read-write",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password)

	poolCfg, err := pgxpool.ParseConfig(connstring)
	if err != nil {
		return nil, xerrors.Errorf("parse cfg: %w", err)
	}

	poolCfg.ConnConfig.TLSConfig = &tls.Config{
		RootCAs:            rootCertPool,
		InsecureSkipVerify: true,
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}

	db := DB{pool: pool, deprecate: cfg.Deprecate}
	if db.deprecate == 0 {
		db.deprecate = defaultDeprecate
	}
	if err := db.prepare(ctx); err != nil {
		return nil, xerrors.Errorf("prepare: %w", err)
	}
	return &db, nil
}

func (db *DB) tx(ctx context.Context, proc func(context.Context, pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return xerrors.Errorf("tx: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := proc(ctx, tx); err != nil {
		return xerrors.Errorf("exec: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return xerrors.Errorf("commit: %w", err)
	}
	return nil
}

func (db *DB) prepare(ctx context.Context) error {
	return db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, eventTable); err != nil {
			return xerrors.Errorf("event schema: %w", err)
		}
		if _, err := tx.Exec(ctx, timerTable); err != nil {
			return xerrors.Errorf("timer schema: %w", err)
		}
		if _, err := tx.Exec(ctx, userTable); err != nil {
			return xerrors.Errorf("user schema: %w", err)
		}
		if _, err := tx.Exec(ctx, stateTable); err != nil {
			return xerrors.Errorf("state schema: %w", err)
		}
		return nil
	})
}

// Get state to properly fetch updates from TGAPI
func (db *DB) GetState(ctx context.Context) (*State, error) {
	var lastOffset int64
	if err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := db.pool.Query(ctx, selectState)
		if err != nil {
			return xerrors.Errorf("select state: %w", err)
		}
		defer rows.Close()
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return xerrors.Errorf("res next: %w", err)
			}
			return nil
		}
		var contents string
		if err := rows.Scan(&lastOffset, &contents); err != nil {
			return xerrors.Errorf("scan: %w", err)
		}
		return nil
	}); err != nil {
		return nil, xerrors.Errorf("tx: %w", err)
	}
	return &State{Offset: lastOffset}, nil
}

// Add events fetched from TGAPI
func (db *DB) RegisterEvents(ctx context.Context, events []Event) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}
	var num int
	err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := db.pool.Query(ctx, selectState)
		if err != nil {
			return xerrors.Errorf("exec: %w", err)
		}
		defer rows.Close()
		var lastOffset int64
		var contents string
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return xerrors.Errorf("res next: %w", err)
			}
		} else if err := rows.Scan(&lastOffset, &contents); err != nil {
			return xerrors.Errorf("scan: %w", err)
		}
		logging.S(ctx).Debugf("Registering %d events from offset %d...", len(events), lastOffset)

		newOffset := lastOffset
		num = 0
		for _, event := range events {
			if event.ID <= lastOffset {
				continue
			}
			logging.S(ctx).Debugf("Add event %q...", event.Payload)
			if _, err := tx.Exec(ctx, insertEvent, event.ID, event.TS, event.UserID, event.Type, event.Payload); err != nil {
				return xerrors.Errorf("insert event: %w", err)
			}
			num++
			if event.ID > newOffset {
				newOffset = event.ID
			}
			if event.UserName != "" {
				logging.S(ctx).Debugf("Add user %q...", event.UserName)
				if _, err := tx.Exec(ctx, insertUser, event.UserID, event.UserName, "", ""); err != nil {
					return xerrors.Errorf("insert user: %w", err)
				}
			}
		}
		logging.S(ctx).Debugf("Setting new offset to %d...", newOffset)
		if _, err := tx.Exec(ctx, updateState, newOffset, contents); err != nil {
			return xerrors.Errorf("update state: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, xerrors.Errorf("register: %w", err)
	}
	return num, nil
}

// Fetch events need to process, filter by busy, mark as busy
func (db *DB) FetchEvents(ctx context.Context) ([]*Event, *User, error) {
	var events []*Event
	users := map[int64]User{}
	var currentUser User
	now := time.Now().Unix()
	deprecate := time.Now().Add(-db.deprecate).Unix()
	logging.S(ctx).Debugf("Fetching events...")
	err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		events = nil
		eventRows, err := db.pool.Query(ctx, listEvents, deprecate)
		if err != nil {
			return xerrors.Errorf("query events: %w", err)
		}
		defer eventRows.Close()
		for eventRows.Next() {
			if err := eventRows.Err(); err != nil {
				return xerrors.Errorf("res next: %w", err)
			}
			var id, ts, userID int64
			var eventType, payload string
			if err := eventRows.Scan(&id, &ts, &userID, &eventType, &payload); err != nil {
				return xerrors.Errorf("scan event: %w", err)
			}
			logging.S(ctx).Debugf("Event from user %d", userID)
			user, ok := users[userID]
			switch {
			case currentUser.ID != 0 && userID != currentUser.ID:
				// another user - skip
				logging.S(ctx).Debugf("Another user, continue")
				continue
			case !ok:
				logging.S(ctx).Debugf("New user %d, fetch", userID)
				userRows, err := db.pool.Query(ctx, selectUser, userID)
				if err != nil {
					return xerrors.Errorf("query user: %w", err)
				}
				defer userRows.Close()
				if userRows.Next() {
					// found
					if err := userRows.Scan(&user.Name, &user.Busy, &user.State, &user.Contents); err != nil {
						return xerrors.Errorf("user scan: %w", err)
					}
					logging.S(ctx).Debugf("User found: %q", user.Name)
				} else if err := userRows.Err(); err != nil {
					return xerrors.Errorf("res next: %w", err)
				} else {
					// no user - skip
					logging.S(ctx).Debugf("None found, skip")
					continue
				}
				logging.S(ctx).Debugf("User is %q", user.Name)
				user.ID = userID
				users[user.ID] = user
				if user.Busy > deprecate {
					// busy - skip
					logging.S(ctx).Debugf("User busy, skip", user.Busy, deprecate)
					continue
				}
			case user.Busy > deprecate:
				// busy - skip
				logging.S(ctx).Debugf("User busy, skip", user.Busy, deprecate)
				continue
			}
			events = append(events, &Event{
				ID:       id,
				TS:       ts,
				UserID:   userID,
				Busy:     now,
				Type:     eventType,
				Payload:  payload,
				UserName: user.Name,
			})
			logging.S(ctx).Debugf("Will process event %q", payload)
			currentUser = user
		}
		for _, event := range events {
			if _, err := tx.Exec(ctx, markEvent, event.ID, event.TS, event.UserID, now); err != nil {
				return xerrors.Errorf("mark event: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, markUser, currentUser.ID, now); err != nil {
			return xerrors.Errorf("mark user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, xerrors.Errorf("tx: %w", err)
	}

	return events, &currentUser, nil
}

// process event
func (db *DB) ProcessEvent(ctx context.Context, user *User, events []*Event) error {
	return db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		logging.S(ctx).Debugf("Processing %d events for user %q...", len(events), user.Name)
		if _, err := tx.Exec(ctx, updateUser, user.ID, user.Name, user.State, user.Contents); err != nil {
			return xerrors.Errorf("insert user: %w", err)
		}
		for _, event := range events {
			if event.Busy < 0 {
				logging.S(ctx).Debugf("Drop event %q", event.Payload)
				if _, err := tx.Exec(ctx, deleteEvent, event.ID, event.TS, event.UserID); err != nil {
					return xerrors.Errorf("delete event: %w", err)
				}
			} else {
				logging.S(ctx).Debugf("Postpone event %q", event.Payload)
				if _, err := tx.Exec(ctx, markEvent, event.ID, event.TS, event.UserID, int64(0)); err != nil {
					return xerrors.Errorf("mark event: %w", err)
				}
			}
		}
		return nil
	})
}

// setup timer
func (db *DB) SetupTimer(ctx context.Context, timer *Timer) error {
	return db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, setTimer, timer.ID, timer.TS, timer.UserID, timer.Repeatable); err != nil {
			return xerrors.Errorf("insert user: %w", err)
		}
		return nil
	})
}

// process timers, generate events, reset or drop timers
func (db *DB) ProcessTimers(ctx context.Context) (int, error) {
	var num int
	err := db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		num = 0
		logging.S(ctx).Debugf("Checking timers...")
		rows, err := db.pool.Query(ctx, fetchTimers, time.Now().Unix())
		if err != nil {
			return xerrors.Errorf("exec: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Err(); err != nil {
				return xerrors.Errorf("res next: %w", err)
			}
			var id, user, ts int64
			var repeatable bool
			if err := rows.Scan(&id, &ts, &user, &repeatable); err != nil {
				return xerrors.Errorf("scan: %w", err)
			}
			logging.S(ctx).Debugf("Timer %d for user %d fired", id, user)
			timerTS := time.Unix(ts, 0)
			if _, err := tx.Exec(ctx, insertEvent, -id, ts, user, EventTypeTimer, strconv.Itoa(int(id))); err != nil {
				return xerrors.Errorf("insert event: %w", err)
			}
			num++
			if repeatable {
				logging.S(ctx).Debugf("Reset timer")
				if _, err := tx.Exec(ctx, setTimer, id, timerTS.Add(24*time.Hour).Unix(), user, true); err != nil {
					return xerrors.Errorf("delete timer: %w", err)
				}
			} else {
				logging.S(ctx).Debugf("Drop timer")
				if _, err := tx.Exec(ctx, deleteTimer, id, user); err != nil {
					return xerrors.Errorf("delete timer: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, xerrors.Errorf("process timers: %w", err)
	}
	return num, nil
}

// Stop timer
func (db *DB) StopTimer(ctx context.Context, timer *Timer) error {
	return db.tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, deleteTimer, timer.ID, timer.UserID); err != nil {
			return xerrors.Errorf("delete timer: %w", err)
		}
		return nil
	})
}

func (db *DB) Close() {
	db.pool.Close()
}
