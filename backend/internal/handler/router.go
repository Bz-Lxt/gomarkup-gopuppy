package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gopuppy/internal/auth"
	"gopuppy/internal/config"
	mw "gopuppy/internal/middleware"
	"gopuppy/internal/service"
	"gopuppy/internal/storage"
	"gopuppy/internal/ws"
)

type Deps struct {
	Cfg      *config.Config
	Pool     *pgxpool.Pool
	Redis    *redis.Client
	Store    storage.Store
	Issuer   *auth.Issuer
	Auth     *service.Auth
	Family   *service.Family
	Pet      *service.Pet
	Checkin  *service.Checkin
	Event    *service.Event
	Reminder *service.Reminder
	Finance  *service.Finance
	Media    *service.Media
	Hub      *ws.Hub
	ForceScan func() error
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(mw.RequestID())
	r.Use(mw.CORS(d.Cfg.CORSOrigins))
	api := &API{d: d}

	r.Get("/healthz", api.Healthz)
	r.Get("/readyz", api.Readyz)
	r.Get("/ws", d.Hub.ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", api.Register)
		r.Post("/auth/login", api.Login)
		r.Post("/auth/refresh", api.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(d.Issuer))
			r.Get("/me", api.Me)
			r.Get("/families", api.ListFamilies)
			r.Post("/families", api.CreateFamily)
			r.Get("/families/{familyID}/members", api.Members)
			r.Post("/families/{familyID}/invites", api.Invite)
			r.Post("/families/join", api.Join)
			r.Delete("/families/{familyID}/members/{userID}", api.RemoveMember)
			r.Get("/families/{familyID}/pets", api.ListPets)
			r.Post("/families/{familyID}/pets", api.CreatePet)
			r.Get("/families/{familyID}/notifications", api.Notifications)
			r.Post("/pets", api.CreatePetBody)
			r.Get("/pets/{petID}", api.GetPet)
			r.Patch("/pets/{petID}", api.UpdatePet)
			r.Delete("/pets/{petID}", api.ArchivePet)
			r.Get("/pets/{petID}/checkins/today", api.TodayCheckins)
			r.Post("/pets/{petID}/checkins", api.ToggleCheckin)
			r.Get("/pets/{petID}/events", api.ListEvents)
			r.Post("/pets/{petID}/events", api.CreateEvent)
			r.Get("/pets/{petID}/reminders", api.ListReminders)
			r.Post("/pets/{petID}/reminders", api.CreateReminder)
			r.Post("/notifications/{logID}/replay", api.Replay)
			r.Post("/admin/reminder-scan", api.ForceScan)
			r.Get("/pets/{petID}/finance", api.Finance)
			r.Post("/pets/{petID}/weights", api.AddWeight)
			r.Post("/pets/{petID}/expenses", api.AddExpense)
			r.Get("/pets/{petID}/media", api.ListMedia)
			r.Post("/pets/{petID}/media", api.UploadMedia)
			r.Get("/media/{mediaID}/file", api.DownloadMedia)
		})
	})
	return r
}

func (a *API) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "time": time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")})
}

func (a *API) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := a.d.Pool.Ping(ctx); err != nil {
		http.Error(w, "db", 503)
		return
	}
	if err := a.d.Redis.Ping(ctx).Err(); err != nil {
		http.Error(w, "redis", 503)
		return
	}
	if err := a.d.Store.Ping(ctx); err != nil {
		http.Error(w, "storage", 503)
		return
	}
	writeJSON(w, 200, map[string]any{"status": "ready", "storage": a.d.Store.Driver()})
}
