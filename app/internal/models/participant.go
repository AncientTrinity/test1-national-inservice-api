package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Participant struct {
	ParticipantID    int64     `json:"participant_id"`
	CourseSessionID  int64     `json:"course_session_id"`
	PersonID         int64     `json:"person_id"`
	Status           string    `json:"status"`
	CreditAwarded    *float64  `json:"credit_awarded,omitempty"`
	Note             *string   `json:"note,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

func CreateParticipant(ctx context.Context, tx pgx.Tx, p *Participant) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `INSERT INTO participant (course_session_id, person_id, status, credit_awarded, note) VALUES ($1,$2,$3,$4,$5) RETURNING participant_id`,
		p.CourseSessionID, p.PersonID, p.Status, p.CreditAwarded, p.Note).Scan(&id)
	return id, err
}

func GetParticipantByID(ctx context.Context, q pgxpool.Pool, id int64) (*Participant, error) {
	row := q.QueryRow(ctx, `SELECT participant_id, course_session_id, person_id, status, credit_awarded, note, created_at FROM participant WHERE participant_id=$1`, id)
	var p Participant
	if err := row.Scan(&p.ParticipantID, &p.CourseSessionID, &p.PersonID, &p.Status, &p.CreditAwarded, &p.Note, &p.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func UpdateParticipant(ctx context.Context, tx pgx.Tx, p *Participant) error {
	_, err := tx.Exec(ctx, `UPDATE participant SET course_session_id=$1, person_id=$2, status=$3, credit_awarded=$4, note=$5 WHERE participant_id=$6`,
		p.CourseSessionID, p.PersonID, p.Status, p.CreditAwarded, p.Note, p.ParticipantID)
	return err
}

func DeleteParticipant(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM participant WHERE participant_id=$1`, id)
	return err
}
