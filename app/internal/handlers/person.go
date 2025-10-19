package handlers

import (
	//"context"
	"encoding/json"
	"net/http"
	"strconv"

	"victortillett.net/test1-national-inservice-api/internal/models"
	//"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/chi/v5"
)


func (h *Handler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	var p models.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	id, err := models.CreatePerson(ctx, tx, &p)
	if err != nil {
		http.Error(w, "insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"person_id": id})
}

func (h *Handler) GetPerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	p, err := models.GetPersonByID(r.Context(), h.db, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) UpdatePerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var p models.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	p.PersonID = id

	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := models.UpdatePerson(ctx, tx, &p); err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeletePerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := models.DeletePerson(ctx, tx, id); err != nil {
		http.Error(w, "delete error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type ListPersonsResponse struct {
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Items   []models.Person `json:"items"`
}

func (h *Handler) ListPersons(w http.ResponseWriter, r *http.Request) {
	// pagination & sorting middleware provide "page","per_page","sort"
	q := r.URL.Query()
	page := 1
	perPage := 20
	if q.Get("page") != "" {
		if v, err := strconv.Atoi(q.Get("page")); err == nil && v > 0 {
			page = v
		}
	}
	if q.Get("per_page") != "" {
		if v, err := strconv.Atoi(q.Get("per_page")); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}
	sort := q.Get("sort")
	if sort == "" {
		sort = "person_id"
	}
	offset := (page - 1) * perPage

	ctx := r.Context()
	rows, err := h.db.Query(ctx, `SELECT person_id, regulation_number, first_name, middle_name, last_name, sex, rank_id, formation_id, posting_id, region_id, phone, email, is_active, created_at, updated_at 
		FROM person ORDER BY `+sort+` LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]models.Person, 0)
	for rows.Next() {
		var p models.Person
		_ = rows.Scan(&p.PersonID, &p.RegulationNumber, &p.FirstName, &p.MiddleName, &p.LastName, &p.Sex, &p.RankID, &p.FormationID, &p.PostingID, &p.RegionID, &p.Phone, &p.Email, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		items = append(items, p)
	}

	var total int
	err = h.db.QueryRow(ctx, `SELECT COUNT(1) FROM person`).Scan(&total)
	if err != nil {
		total = len(items)
	}

	resp := ListPersonsResponse{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	}
	json.NewEncoder(w).Encode(resp)
}
