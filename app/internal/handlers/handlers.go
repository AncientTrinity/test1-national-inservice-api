package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"victortillett.net/test1-national-inservice-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{db: pool}
}

/* -------- Courses CRUD -------- */

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var c models.Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)

	id, err := models.CreateCourse(ctx, tx, &c)
	if err != nil { http.Error(w, "insert error: "+err.Error(), http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"course_id": id})
}

func (h *Handler) GetCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	c, err := models.GetCourseByID(r.Context(), h.db, id)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	if c == nil { http.Error(w, "not found", http.StatusNotFound); return }
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var c models.Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil { http.Error(w, "invalid body", http.StatusBadRequest); return }
	c.CourseID = id
	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)
	if err := models.UpdateCourse(ctx, tx, &c); err != nil { http.Error(w, "update error", http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)
	if err := models.DeleteCourse(ctx, tx, id); err != nil { http.Error(w, "delete error", http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := 1; perPage := 20
	if v := q.Get("page"); v != "" { if n, err := strconv.Atoi(v); err == nil && n>0 { page = n } }
	if v := q.Get("per_page"); v != "" { if n, err := strconv.Atoi(v); err == nil && n>0 && n<=100 { perPage = n } }
	sort := q.Get("sort"); if sort=="" { sort = "course_id" }
	offset := (page-1)*perPage

	ctx := r.Context()
	rows, err := h.db.Query(ctx, `SELECT course_id, code, title, description, category, credit_hours, is_course_active, rating, created_at FROM courses ORDER BY `+sort+` LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer rows.Close()

	items := []models.Course{}
	for rows.Next() {
		var c models.Course
		_ = rows.Scan(&c.CourseID, &c.Code, &c.Title, &c.Description, &c.Category, &c.CreditHours, &c.IsActive, &c.Rating, &c.CreatedAt)
		items = append(items, c)
	}
	var total int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(1) FROM courses`).Scan(&total)
	resp := map[string]interface{}{"total": total, "page": page, "per_page": perPage, "items": items}
	json.NewEncoder(w).Encode(resp)
}

/* -------- Participant CRUD -------- */

func (h *Handler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var p models.Participant
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "invalid body", http.StatusBadRequest); return }
	ctx := r.Context()
	tx, err := h.db.Begin(ctx); if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)
	id, err := models.CreateParticipant(ctx, tx, &p)
	if err != nil { http.Error(w, "insert error", http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"participant_id": id})
}

func (h *Handler) GetParticipant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id"); id, _ := strconv.ParseInt(idStr, 10, 64)
	p, err := models.GetParticipantByID(r.Context(), h.db, id)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	if p == nil { http.Error(w, "not found", http.StatusNotFound); return }
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) UpdateParticipant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id"); id, _ := strconv.ParseInt(idStr, 10, 64)
	var p models.Participant
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "invalid body", http.StatusBadRequest); return }
	p.ParticipantID = id
	ctx := r.Context(); tx, err := h.db.Begin(ctx); if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)
	if err := models.UpdateParticipant(ctx, tx, &p); err != nil { http.Error(w, "update error", http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id"); id, _ := strconv.ParseInt(idStr, 10, 64)
	ctx := r.Context(); tx, err := h.db.Begin(ctx); if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer tx.Rollback(ctx)
	if err := models.DeleteParticipant(ctx, tx, id); err != nil { http.Error(w, "delete error", http.StatusInternalServerError); return }
	if err := tx.Commit(ctx); err != nil { http.Error(w, "commit error", http.StatusInternalServerError); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := 1; perPage := 20
	if v := q.Get("page"); v != "" { if n, err := strconv.Atoi(v); err == nil && n>0 { page=n } }
	if v := q.Get("per_page"); v != "" { if n, err := strconv.Atoi(v); err == nil && n>0 && n<=100 { perPage=n } }
	offset := (page-1)*perPage
	ctx := r.Context()
	rows, err := h.db.Query(ctx, `SELECT participant_id, course_session_id, person_id, status, credit_awarded, note, created_at FROM participant LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil { http.Error(w, "db error", http.StatusInternalServerError); return }
	defer rows.Close()
	items := []models.Participant{}
	for rows.Next() {
		var p models.Participant
		_ = rows.Scan(&p.ParticipantID, &p.CourseSessionID, &p.PersonID, &p.Status, &p.CreditAwarded, &p.Note, &p.CreatedAt)
		items = append(items, p)
	}
	var total int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(1) FROM participant`).Scan(&total)
	resp := map[string]interface{}{"total": total, "page": page, "per_page": perPage, "items": items}
	json.NewEncoder(w).Encode(resp)
}
