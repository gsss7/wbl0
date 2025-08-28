package http

import (
	"encoding/json"
	"net/http"

	"wbl0/internal/cache"
	"wbl0/internal/models"
	"wbl0/internal/repo"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Repo  *repo.Repository
	Cache *cache.Cache[models.Order]
}

type errResp struct {
	Error string `json:"error"`
}

func (a *API) Register(r *chi.Mux) {
	r.Get("/api/healthz", a.health)
	r.Get("/api/order/{id}", a.getOrder)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" || len(id) > 64 {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid order id"})
		return
	}

	if v, ok := a.Cache.Get(id); ok {
		writeJSON(w, http.StatusOK, v)
		return
	}

	ord, err := a.Repo.GetOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errResp{Error: "not found"})
		return
	}
	a.Cache.Set(id, *ord)
	writeJSON(w, http.StatusOK, ord)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
