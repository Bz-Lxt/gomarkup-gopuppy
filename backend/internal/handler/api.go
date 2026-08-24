package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gopuppy/internal/domain"
	"gopuppy/internal/httputil"
	mw "gopuppy/internal/middleware"
	"gopuppy/internal/service"
	"gopuppy/internal/storage"
)

type API struct{ d Deps }

func rid(r *http.Request) string { return mw.RequestIDFrom(r.Context()) }

func uid(r *http.Request) uuid.UUID {
	id, _ := mw.UserIDFrom(r.Context())
	return id
}

func parseID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var in service.RegisterInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	u, tok, err := a.d.Auth.Register(r.Context(), in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), map[string]any{"user": u, "tokens": tok})
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	u, tok, err := a.d.Auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), map[string]any{"user": u, "tokens": tok})
}

func (a *API) Refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	tok, err := a.d.Auth.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), tok)
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	u, err := a.d.Auth.Me(r.Context(), uid(r))
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), u)
}

func (a *API) ListFamilies(w http.ResponseWriter, r *http.Request) {
	list, err := a.d.Family.List(r.Context(), uid(r))
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	f, err := a.d.Family.Create(r.Context(), uid(r), in.Name)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), f)
}

func (a *API) Members(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Family.Members(r.Context(), uid(r), fid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) Invite(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in struct {
		Role domain.Role `json:"role"`
	}
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	inv, err := a.d.Family.Invite(r.Context(), uid(r), fid, in.Role)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), inv)
}

func (a *API) Join(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code string `json:"code"`
	}
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	f, err := a.d.Family.Join(r.Context(), uid(r), in.Code)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), f)
}

func (a *API) RemoveMember(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	tid, err := parseID(r, "userID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	if err := a.d.Family.Remove(r.Context(), uid(r), fid, tid); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), map[string]string{"status": "removed"})
}

func (a *API) ListPets(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Pet.List(r.Context(), uid(r), fid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) CreatePet(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.PetInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	in.FamilyID = fid
	p, err := a.d.Pet.Create(r.Context(), uid(r), in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), p)
}

func (a *API) CreatePetBody(w http.ResponseWriter, r *http.Request) {
	var in service.PetInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	p, err := a.d.Pet.Create(r.Context(), uid(r), in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), p)
}

func (a *API) GetPet(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	p, err := a.d.Pet.Get(r.Context(), uid(r), pid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), p)
}

func (a *API) UpdatePet(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.PetInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	p, err := a.d.Pet.Update(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), p)
}

func (a *API) ArchivePet(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	if err := a.d.Pet.Archive(r.Context(), uid(r), pid); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), map[string]string{"status": "archived"})
}

func (a *API) TodayCheckins(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Checkin.Today(r.Context(), uid(r), pid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) ToggleCheckin(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.CheckinInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	list, err := a.d.Checkin.Toggle(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) ListEvents(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	list, err := a.d.Event.List(r.Context(), uid(r), pid, r.URL.Query().Get("category"), year)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) CreateEvent(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.EventInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	e, err := a.d.Event.Create(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), e)
}

func (a *API) ListReminders(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Reminder.List(r.Context(), uid(r), pid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) CreateReminder(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.RuleInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	rule, err := a.d.Reminder.Create(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), rule)
}

func (a *API) Notifications(w http.ResponseWriter, r *http.Request) {
	fid, err := parseID(r, "familyID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Reminder.Logs(r.Context(), uid(r), fid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) Replay(w http.ResponseWriter, r *http.Request) {
	lid, err := parseID(r, "logID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	if err := a.d.Reminder.Replay(r.Context(), uid(r), lid); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), map[string]string{"status": "replayed"})
}

func (a *API) ForceScan(w http.ResponseWriter, r *http.Request) {
	if a.d.ForceScan == nil {
		httputil.OK(w, rid(r), map[string]string{"status": "noop"})
		return
	}
	if err := a.d.ForceScan(); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), map[string]string{"status": "scanned"})
}

func (a *API) Finance(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	sum, err := a.d.Finance.Summary(r.Context(), uid(r), pid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), sum)
}

func (a *API) AddWeight(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.WeightInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	wrec, err := a.d.Finance.AddWeight(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), wrec)
}

func (a *API) AddExpense(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	var in service.ExpenseInput
	if err := httputil.DecodeJSON(r, &in); err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	e, err := a.d.Finance.AddExpense(r.Context(), uid(r), pid, in)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), e)
}

func (a *API) ListMedia(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	list, err := a.d.Media.List(r.Context(), uid(r), pid, r.URL.Query().Get("kind"))
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.OK(w, rid(r), list)
}

func (a *API) UploadMedia(w http.ResponseWriter, r *http.Request) {
	pid, err := parseID(r, "petID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		httputil.Error(w, rid(r), domain.ErrTooLarge)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	defer file.Close()
	kind := domain.MediaKind(r.FormValue("kind"))
	m, err := a.d.Media.Upload(r.Context(), uid(r), pid, kind, hdr.Filename, file)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	httputil.Created(w, rid(r), m)
}

func (a *API) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	mid, err := parseID(r, "mediaID")
	if err != nil {
		httputil.Error(w, rid(r), domain.ErrValidation)
		return
	}
	m, rc, err := a.d.Media.Open(r.Context(), uid(r), mid)
	if err != nil {
		httputil.Error(w, rid(r), err)
		return
	}
	defer rc.Close()
	_ = storage.MaxFileBytes
	w.Header().Set("Content-Type", m.MIME)
	w.Header().Set("Content-Disposition", `inline; filename="`+m.Filename+`"`)
	_, _ = io.Copy(w, rc)
}
