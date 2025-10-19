package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseSession struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location"`
	RegionID    int64     `json:"region_id"`
	FormationID int64     `json:"formation_id"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
}

type CourseSessionModel struct {
	DB *pgxpool.Pool
}

// Create a new course session
func (m *CourseSessionModel) Insert(ctx context.Context, cs *CourseSession) error {
	query := `
		INSERT INTO course_session (course_id, start_date, end_date, location, region_id, formation_id, capacity, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING course_sess_id, created_at`

	return m.DB.QueryRow(ctx, query,
		cs.CourseID,
		cs.StartDate,
		cs.EndDate,
		cs.Location,
		cs.RegionID,
		cs.FormationID,
		cs.Capacity,
	).Scan(&cs.ID, &cs.CreatedAt)
}

// Get a single course session
func (m *CourseSessionModel) GetByID(ctx context.Context, id int64) (*CourseSession, error) {
	query := `
		SELECT course_sess_id, course_id, start_date, end_date, location, region_id, formation_id, capacity, created_at
		FROM course_session
		WHERE course_sess_id = $1`

	var cs CourseSession
	err := m.DB.QueryRow(ctx, query, id).Scan(
		&cs.ID,
		&cs.CourseID,
		&cs.StartDate,
		&cs.EndDate,
		&cs.Location,
		&cs.RegionID,
		&cs.FormationID,
		&cs.Capacity,
		&cs.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// Update a course session
func (m *CourseSessionModel) Update(ctx context.Context, cs *CourseSession) error {
	query := `
		UPDATE course_session
		SET course_id = $1, start_date = $2, end_date = $3, location = $4,
		    region_id = $5, formation_id = $6, capacity = $7
		WHERE course_sess_id = $8`

	_, err := m.DB.Exec(ctx, query,
		cs.CourseID,
		cs.StartDate,
		cs.EndDate,
		cs.Location,
		cs.RegionID,
		cs.FormationID,
		cs.Capacity,
		cs.ID,
	)
	return err
}

// Delete a course session
func (m *CourseSessionModel) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM course_session WHERE course_sess_id = $1`
	cmdTag, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("course session not found")
	}
	return nil
}

// List with pagination and sorting
func (m *CourseSessionModel) List(ctx context.Context, limit, offset int, sortBy string) ([]*CourseSession, error) {
	if sortBy == "" {
		sortBy = "created_at"
	}
	query := `
		SELECT course_sess_id, course_id, start_date, end_date, location, region_id, formation_id, capacity, created_at
		FROM course_session
		ORDER BY ` + sortBy + ` DESC
		LIMIT $1 OFFSET $2`

	rows, err := m.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*CourseSession
	for rows.Next() {
		var cs CourseSession
		err := rows.Scan(
			&cs.ID,
			&cs.CourseID,
			&cs.StartDate,
			&cs.EndDate,
			&cs.Location,
			&cs.RegionID,
			&cs.FormationID,
			&cs.Capacity,
			&cs.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &cs)
	}
	return sessions, nil
}
