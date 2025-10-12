package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Person struct {
	PersonID         int64      `json:"person_id"`
	RegulationNumber *string    `json:"regulation_number,omitempty"`
	FirstName        string     `json:"first_name"`
	MiddleName       *string    `json:"middle_name,omitempty"`
	LastName         string     `json:"last_name"`
	Sex              *string    `json:"sex,omitempty"`
	RankID           *int64     `json:"rank_id,omitempty"`
	FormationID      *int64     `json:"formation_id,omitempty"`
	PostingID        *int64     `json:"posting_id,omitempty"`
	RegionID         *int64     `json:"region_id,omitempty"`
	Phone            *string    `json:"phone,omitempty"`
	Email            *string    `json:"email,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

// CreatePerson inserts a person and returns new ID
func CreatePerson(ctx context.Context, tx pgx.Tx, p *Person) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO person (regulation_number, first_name, middle_name, last_name, sex, rank_id, formation_id, posting_id, region_id, phone, email, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING person_id
	`, p.RegulationNumber, p.FirstName, p.MiddleName, p.LastName, p.Sex, p.RankID, p.FormationID, p.PostingID, p.RegionID, p.Phone, p.Email, p.IsActive).Scan(&id)
	return id, err
}

func GetPersonByID(ctx context.Context, conn pgx.Queryer, id int64) (*Person, error) {
	row := conn.QueryRow(ctx, `SELECT person_id, regulation_number, first_name, middle_name, last_name, sex, rank_id, formation_id, posting_id, region_id, phone, email, is_active, created_at, updated_at FROM person WHERE person_id=$1`, id)
	var p Person
	err := row.Scan(&p.PersonID, &p.RegulationNumber, &p.FirstName, &p.MiddleName, &p.LastName, &p.Sex, &p.RankID, &p.FormationID, &p.PostingID, &p.RegionID, &p.Phone, &p.Email, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdatePerson(ctx context.Context, tx pgx.Tx, p *Person) error {
	_, err := tx.Exec(ctx, `UPDATE person SET regulation_number=$1, first_name=$2, middle_name=$3, last_name=$4, sex=$5, rank_id=$6, formation_id=$7, posting_id=$8, region_id=$9, phone=$10, email=$11, is_active=$12, updated_at=NOW() WHERE person_id=$13`,
		p.RegulationNumber, p.FirstName, p.MiddleName, p.LastName, p.Sex, p.RankID, p.FormationID, p.PostingID, p.RegionID, p.Phone, p.Email, p.IsActive, p.PersonID)
	return err
}

func DeletePerson(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM person WHERE person_id=$1`, id)
	return err
}
