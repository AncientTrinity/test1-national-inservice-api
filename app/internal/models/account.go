package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	AccountID    int64     `json:"account_id"`
	PersonID     *int64    `json:"person_id,omitempty"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateAccount inserts a new account and returns the new ID.
func CreateAccount(ctx context.Context, pool *pgxpool.Pool, a *Account) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO account (person_id, email, password_hash, role, is_active)
         VALUES ($1,$2,$3,$4,$5) RETURNING account_id`,
		a.PersonID, a.Email, a.PasswordHash, a.Role, a.IsActive,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetAccountByEmail fetches an account by email.
func GetAccountByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*Account, error) {
	row := pool.QueryRow(ctx,
		`SELECT account_id, person_id, email, password_hash, role, is_active, created_at
         FROM account WHERE email=$1`, email)

	var a Account
	if err := row.Scan(&a.AccountID, &a.PersonID, &a.Email,
		&a.PasswordHash, &a.Role, &a.IsActive, &a.CreatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAccountByID fetches an account by its ID.
func GetAccountByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*Account, error) {
	row := pool.QueryRow(ctx,
		`SELECT account_id, person_id, email, password_hash, role, is_active, created_at
         FROM account WHERE account_id=$1`, id)

	var a Account
	if err := row.Scan(&a.AccountID, &a.PersonID, &a.Email,
		&a.PasswordHash, &a.Role, &a.IsActive, &a.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// UpdateAccount updates account fields.
func UpdateAccount(ctx context.Context, tx pgx.Tx, a *Account) error {
	_, err := tx.Exec(ctx,
		`UPDATE account
         SET person_id=$1, email=$2, password_hash=$3, role=$4, is_active=$5
         WHERE account_id=$6`,
		a.PersonID, a.Email, a.PasswordHash, a.Role, a.IsActive, a.AccountID)
	return err
}

// DeleteAccount deletes an account by ID.
func DeleteAccount(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx, `DELETE FROM account WHERE account_id=$1`, id)
	return err
}
