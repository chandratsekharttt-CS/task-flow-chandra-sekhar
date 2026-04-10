package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/models"
)

// UserRepository handles all user-related database operations.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user and returns the created user.
func (r *UserRepository) Create(ctx context.Context, name, email, hashedPassword string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, email, password, created_at`,
		name, email, hashedPassword,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail looks up a user by email. Returns nil if not found.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID looks up a user by ID. Returns nil if not found.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, password, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if a user with the given email already exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	).Scan(&exists)
	return exists, err
}

// CountAll returns the total number of users (used for seed check).
func (r *UserRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// ListAll returns all users, excluding passwords.
func (r *UserRepository) ListAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, email, created_at FROM users ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
