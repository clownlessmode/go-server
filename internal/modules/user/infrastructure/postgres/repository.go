package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"project/internal/modules/user/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	created := &domain.User{}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (login, password, role, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, login, password, role, is_active, created_at, updated_at
	`, user.Login, user.Password, string(user.Role), user.IsActive).Scan(
		&created.ID,
		&created.Login,
		&created.Password,
		&created.Role,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (r *Repository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, login, password, role, is_active, created_at, updated_at
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users rows: %w", err)
	}

	return users, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, login, password, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *Repository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	user := &domain.User{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, login, password, role, is_active, created_at, updated_at
		FROM users
		WHERE login = $1
	`, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return user, nil
}

func (r *Repository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	updated := &domain.User{}

	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET login = $1,
			password = $2,
			role = $3,
			is_active = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING id, login, password, role, is_active, created_at, updated_at
	`, user.Login, user.Password, string(user.Role), user.IsActive, user.ID).Scan(
		&updated.ID,
		&updated.Login,
		&updated.Password,
		&updated.Role,
		&updated.IsActive,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user affected rows: %w", err)
	}
	if affected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (*domain.User, error) {
	user := &domain.User{}
	if err := scanner.Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return user, nil
}
