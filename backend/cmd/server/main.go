package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopuppy/internal/auth"
	"gopuppy/internal/config"
	"gopuppy/internal/handler"
	"gopuppy/internal/logger"
	"gopuppy/internal/notifier"
	"gopuppy/internal/reminder"
	"gopuppy/internal/repo"
	"gopuppy/internal/service"
	"gopuppy/internal/storage"
	"gopuppy/internal/ws"
	"gopuppy/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := logger.New(cfg.LogLevel, cfg.AppEnv)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repo.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := repo.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Error("redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	if err := repo.Migrate(ctx, pool, migrations.FS); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}
	if err := repo.Seed(ctx, pool); err != nil {
		log.Error("seed", "err", err)
		os.Exit(1)
	}

	store, err := storage.New(cfg)
	if err != nil {
		log.Error("storage", "err", err)
		os.Exit(1)
	}

	issuer := &auth.Issuer{Secret: []byte(cfg.JWTSecret), AccessTTL: cfg.AccessTTL, RefreshTTL: cfg.RefreshTTL}
	users := &repo.Users{Pool: pool}
	families := &repo.Families{Pool: pool}
	pets := &repo.Pets{Pool: pool}
	checkins := &repo.Checkins{Pool: pool}
	events := &repo.Events{Pool: pool}
	reminders := &repo.Reminders{Pool: pool}
	finance := &repo.Finance{Pool: pool}
	mediaRepo := &repo.Media{Pool: pool}
	hub := ws.New(log, issuer, families)
	famSvc := &service.Family{Families: families, Pets: pets}
	remSvc := &service.Reminder{
		Rules: reminders, Family: famSvc, Pets: pets, Users: users, Families: families,
		Sender: notifier.New(cfg, mediaRepo),
	}
	sched := &reminder.Scheduler{
		Log: log, ScanHour: cfg.ReminderScanHour, Tick: time.Duration(cfg.ReminderTickSec) * time.Second, Scanner: remSvc,
	}
	sched.Start(ctx)

	h := handler.New(handler.Deps{
		Cfg: cfg, Pool: pool, Redis: rdb, Store: store, Issuer: issuer,
		Auth:     &service.Auth{Users: users, Issuer: issuer},
		Family:   famSvc,
		Pet:      &service.Pet{Pets: pets, Family: famSvc},
		Checkin:  &service.Checkin{Repo: checkins, Family: famSvc, Hub: hub},
		Event:    &service.Event{Events: events, Reminders: reminders, Family: famSvc, Hub: hub},
		Reminder: remSvc,
		Finance:  &service.Finance{Repo: finance, Family: famSvc},
		Media:    &service.Media{Repo: mediaRepo, Family: famSvc, Store: store},
		Hub:      hub,
		ForceScan: func() error { return sched.ForceScan(context.Background()) },
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	go func() {
		log.Info("listen", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	sh, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(sh)
}
