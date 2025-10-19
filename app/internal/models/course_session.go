package models

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// CourseSession represents an instance of a Course being offered in a specific time frame or region.
type CourseSession struct {
    CourseSessionID int64      `json:"course_session_id"`
    CourseID        int64      `json:"course_id"`
    StartDate       time.Time  `json:"start_date"`
    EndDate         *time.Time `json:"end_date,omitempty"`
    Location        *string    `json:"location,omitempty"`
    RegionID        *int64     `json:"region_id,omitempty"`
    FormationID     *int64     `json:"formation_id,omitempty"`
    Capacity        *int       `json:"capacity,omitempty"`
    CreatedAt       time.Time  `json:"created_at"`
}

//
// ─── CRUD OPERATIONS ─────────────────────────────────────────────────────────────
//

// CreateCourseSession inserts a new course session into the database and returns its ID.
func CreateCourseSession(ctx context.Context, tx pgx.Tx, cs *CourseSession) (int64, error) {
    var id int64
    err := tx.QueryRow(ctx, `
        INSERT INTO course_sessions (course_id, start_date, end_date, location, region_id, formation_id, capacity)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING course_session_id`,
        cs.CourseID, cs.StartDate, cs.EndDate, cs.Location, cs.RegionID, cs.FormationID, cs.Capacity,
    ).Scan(&id)
    return id, err
}

// GetCourseSessionByID retrieves a course session by its ID.
func GetCourseSessionByID(ctx context.Context, db *pgxpool.Pool, id int64) (*CourseSession, error) {
    row := db.QueryRow(ctx, `
        SELECT course_session_id, course_id, start_date, end_date, location, region_id, formation_id, capacity, created_at
        FROM course_sessions
        WHERE course_session_id=$1`, id)

    var cs CourseSession
    if err := row.Scan(
        &cs.CourseSessionID,
        &cs.CourseID,
        &cs.StartDate,
        &cs.EndDate,
        &cs.Location,
        &cs.RegionID,
        &cs.FormationID,
        &cs.Capacity,
        &cs.CreatedAt,
    ); err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &cs, nil
}

// UpdateCourseSession updates an existing course session’s data.
func UpdateCourseSession(ctx context.Context, tx pgx.Tx, cs *CourseSession) error {
    _, err := tx.Exec(ctx, `
        UPDATE course_sessions
        SET course_id=$1, start_date=$2, end_date=$3, location=$4,
            region_id=$5, formation_id=$6, capacity=$7
        WHERE course_session_id=$8`,
        cs.CourseID, cs.StartDate, cs.EndDate, cs.Location,
        cs.RegionID, cs.FormationID, cs.Capacity, cs.CourseSessionID,
    )
    return err
}

// DeleteCourseSession removes a course session by its ID.
func DeleteCourseSession(ctx context.Context, tx pgx.Tx, id int64) error {
    _, err := tx.Exec(ctx, `DELETE FROM course_sessions WHERE course_session_id=$1`, id)
    return err
}

// GetAllCourseSessions retrieves all course sessions.
func GetAllCourseSessions(ctx context.Context, db *pgxpool.Pool) ([]*CourseSession, error) {
    rows, err := db.Query(ctx, `
        SELECT course_session_id, course_id, start_date, end_date, location,
               region_id, formation_id, capacity, created_at
        FROM course_sessions
        ORDER BY start_date DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sessions []*CourseSession
    for rows.Next() {
        var cs CourseSession
        if err := rows.Scan(
            &cs.CourseSessionID,
            &cs.CourseID,
            &cs.StartDate,
            &cs.EndDate,
            &cs.Location,
            &cs.RegionID,
            &cs.FormationID,
            &cs.Capacity,
            &cs.CreatedAt,
        ); err != nil {
            return nil, err
        }
        sessions = append(sessions, &cs)
    }
    return sessions, rows.Err()
}
