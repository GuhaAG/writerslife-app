package db

import (
	"context"
	"fmt"

	"github.com/guhaag/writerslife-books-api/model"
)

const userColumns = `id, username, email, password_hash, bio, avatar_url, created_at`

func (s *Store) CreateUser(ctx context.Context, user *model.User) error {
	return s.pool.QueryRow(ctx, `
		INSERT INTO users (id, username, email, password_hash, bio, avatar_url, created_at)
		VALUES ('u' || nextval('user_id_sequence'), $1, $2, $3, $4, $5, $6)
		RETURNING `+userColumns,
		user.Username, user.Email, user.PasswordHash, user.Bio, user.AvatarURL, user.CreatedAt,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Bio, &user.AvatarURL, &user.CreatedAt)
}

func (s *Store) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	return s.findUser(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

func (s *Store) FindUserByName(ctx context.Context, username string) (*model.User, error) {
	return s.findUser(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username)
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.findUser(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email)
}

func (s *Store) findUser(ctx context.Context, query, value string) (*model.User, error) {
	user := &model.User{}
	err := s.pool.QueryRow(ctx, query, value).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Bio, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}
