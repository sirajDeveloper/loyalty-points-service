package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	domainerrors "github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/errors"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/model"
	"github.com/sirajDeveloper/loyalty-points-service/internal/user-service/domain/repository"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (login, password_hash, first_name, last_name, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.pool.QueryRow(ctx, query, user.Login, user.PasswordHash, user.FirstName, user.LastName, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.ErrLoginAlreadyExists
		}
		return err
	}
	return nil
}

func (r *userRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `SELECT id, login, password_hash, first_name, last_name, created_at FROM users WHERE login = $1`
	user := &model.User{}
	err := r.pool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, login).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

