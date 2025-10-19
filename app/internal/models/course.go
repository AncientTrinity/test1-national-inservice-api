package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Course struct {
	CourseID    int64     `json:"course_id"`
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Category    *string   `json:"category,omitempty"`
	CreditHours *float64  `json:"credit_hours,omitempty"`
	IsActive    bool      `json:"is_course_active"`
	Rating      *float64  `json:"rating,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func CreateCourse(ctx context.Context, tx pgx.Tx, c *Course) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `INSERT INTO courses (code, title, description, category, credit_hours, is_course_active, rating)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING course_id`,
		c.Code, c.Title, c.Description, c.Category, c.CreditHours, c.IsActive, c.Rating).Scan(&id)
	return id, err
}

func GetCourseByID(ctx context.Context, q pgx.Queryer, id int64) (*Course, error) {
	row := q.QueryRow(ctx, `SELECT course_id, code, title, description, category, credit_hours, is_course_active, rating, created_at FROM courses WHERE course_id=$1`, id)
	var c Course
	if err := row.Scan(&c.CourseID, &c.Code, &c.Title, &c.Description, &c.Category, &c.CreditHours, &c.IsActive, &c.Rating, &c.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func UpdateCourse(ctx context.Context, tx pgx.Tx, c *Course) error {
	_, err := tx.Exec(ctx, `UPDATE courses SET code=$1, title=$2, description=$3, category=$4, credit_hours=$5, is_course_active=$6, rating=$7 WHERE course_id=$8`,
		c.Code, c.Title, c.Description, c.Category, c.CreditHours, c.IsActive, c.Rating, c.CourseID)
	return err
}

func DeleteCourse(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM courses WHERE course_id=$1`, id)
	return err
}
