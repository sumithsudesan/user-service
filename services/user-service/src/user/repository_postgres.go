package user

import (
	"context"
	"errors"
	"time"

	"github.com/sumithsudesan/pkg/database"
	"github.com/sumithsudesan/pkg/logger"
)

// represent Postgres repository
type postgresRepository struct {
	db  database.DB
	log logger.Logger
}

// Create new Postgres repository instance
func NewPostgresRepository(db database.DB,
	log logger.Logger) Repository {
	return &postgresRepository{db: db,
		log: log}
}

// Craete User
func (r *postgresRepository) Create(u *User) error {

	// Insert user into the database
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO users (id, name, email, status, created_at, updated_at, version)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Name, u.Email, u.Status, u.CreatedAt, u.UpdatedAt, u.Version,
	)
	if err != nil {
		if errors.Is(err, database.ErrConflict) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// Used to get specific user
func (r *postgresRepository) Get(id string) (*User, error) {
	u := &User{}
	// Select queery
	err := r.db.QueryRow(context.Background(),
		`SELECT id, name, email, status, created_at, updated_at, version
         FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.Version)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// List user
func (r *postgresRepository) List() ([]*User, error) {

	// select query
	rows, err := r.db.Query(context.Background(),
		`SELECT id, name, email, status, created_at, updated_at, version
         FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.Version); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// used to update user
func (r *postgresRepository) Update(u *User) error {
	// Updatequery
	result, err := r.db.Exec(context.Background(),
		`UPDATE users
         SET name=$2, email=$3, status=$4, updated_at=$5, version=version+1
         WHERE id=$1 AND version=$6`,
		u.ID, u.Name, u.Email, u.Status, u.UpdatedAt, u.Version,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		if _, err := r.Get(u.ID); err != nil {
			return ErrUserNotFound
		}
		return ErrVersionMismatch
	}
	return nil
}

// used to delete user
func (r *postgresRepository) Delete(id string) error {
	// delete query
	result, err := r.db.Exec(context.Background(),
		`DELETE FROM users WHERE id=$1`, id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// scanTime handles timezone-aware timestamps from PostgreSQL.
var _ = time.UTC
