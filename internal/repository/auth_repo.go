package repository

import (
	"context"
	"fmt"

	"github.com/DobryySoul/Calc-service/internal/http/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pg *pgxpool.Pool
}

func NewAuthRepo(pg *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pg: pg}
}

func (r *AuthRepository) Register(ctx context.Context, user *models.User) error {
	const query = `INSERT INTO users (email, pass_hash) VALUES ($1, $2)`

	_, err := r.pg.Exec(
		ctx,
		query,
		user.Email,
		user.Password,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query and register new user: %w", err)
	}

	return nil
}

func (r *AuthRepository) Login(ctx context.Context, email, password string) (*models.User, error) {
	var user models.User

	const query = `SELECT id, email, pass_hash FROM users WHERE email = $1`

	err := r.pg.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query and login user: %w", err)
	}

	return &user, nil
}
