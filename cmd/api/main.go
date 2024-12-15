package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/baldisbk/tgbot/pkg/logging"
	"github.com/baldisbk/tgbot/pkg/tgapi"
	"go.uber.org/zap"

	"github.com/baldisbk/tgeasybot/internal/config"
	"github.com/baldisbk/tgeasybot/internal/orm"
	"github.com/baldisbk/tgeasybot/internal/poller"
	"github.com/baldisbk/tgeasybot/internal/sm"
	"github.com/baldisbk/tgeasybot/internal/timer"
)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Logger: %#v", err)
		os.Exit(1)
	}
	ctx = logging.WithLogger(ctx, logger)
	defer func() {
		logging.S(ctx).Debugf("Sync logger")
		logger.Sync()
	}()

	logging.S(ctx).Debugf("Parsing config...")
	config, err := config.ParseConfig()
	if err != nil {
		logging.S(ctx).Errorf("Read config: %#v", err)
		os.Exit(1)
	}
	logging.S(ctx).Debugf("Config: %#v", config)

	logging.S(ctx).Debugf("Creating client...")
	client, err := tgapi.NewClient(ctx, config.ApiConfig)
	if err != nil {
		logging.S(ctx).Errorf("TG client: %#v", err)
		os.Exit(1)
	}

	logging.S(ctx).Debugf("Init database...")
	db, err := orm.NewDB(ctx, config.DBConfig)
	if err != nil {
		logging.S(ctx).Errorf("DB client: %#v", err)
		os.Exit(1)
	}
	defer func() {
		logging.S(ctx).Debugf("Stopping DB...")
		db.Close()
		logging.S(ctx).Debugf("DB stopped")
	}()

	logging.S(ctx).Debugf("Starting timers...")
	tim := timer.NewTimerPoller(ctx, config.TimerConfig, db)
	defer func() {
		logging.S(ctx).Debugf("Stopping timer...")
		defer tim.Shutdown()
		logging.S(ctx).Debugf("Timer stopped")
	}()

	logging.S(ctx).Debugf("Starting poller...")
	poll := poller.NewPoller(ctx, config.PollerConfig, client, db)
	defer func() {
		logging.S(ctx).Debugf("Stopping poller...")
		defer poll.Shutdown()
		logging.S(ctx).Debugf("Poller stopped")
	}()

	logging.S(ctx).Debugf("Starting statemachine...")
	sm := sm.NewWorkerPool(ctx, config.WorkerConfig, client, db)
	defer func() {
		logging.S(ctx).Debugf("Stopping statemachine...")
		defer sm.Shutdown()
		logging.S(ctx).Debugf("Statemachine stopped")
	}()

	logging.S(ctx).Debugf("Bot started")

	<-signals

	logging.S(ctx).Debugf("Bot stopped")
}
