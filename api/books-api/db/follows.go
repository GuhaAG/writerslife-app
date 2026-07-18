package db

import (
	"context"
	"fmt"
	"github.com/guhaag/writerslife-books-api/model"
)

func (s *Store) Follow(ctx context.Context, userID, fictionID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO follows (user_id, fiction_id, created_at) VALUES ($1,$2,now()) ON CONFLICT DO NOTHING`, userID, fictionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		if _, err = tx.Exec(ctx, `UPDATE fictions SET follower_count=follower_count+1 WHERE id=$1`, fictionID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (s *Store) Unfollow(ctx context.Context, userID, fictionID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM follows WHERE user_id=$1 AND fiction_id=$2`, userID, fictionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		if _, err = tx.Exec(ctx, `UPDATE fictions SET follower_count=GREATEST(follower_count-1,0) WHERE id=$1`, fictionID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (s *Store) IsFollowing(ctx context.Context, userID, fictionID string) (bool, error) {
	var following bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM follows WHERE user_id=$1 AND fiction_id=$2)`, userID, fictionID).Scan(&following)
	return following, err
}
func (s *Store) ListFollowedFictions(ctx context.Context, userID string) ([]*model.Fiction, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+fictionColumns+` FROM fictions WHERE id IN (SELECT fiction_id FROM follows WHERE user_id = $1) ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list follows: %w", err)
	}
	defer rows.Close()
	return scanFictions(rows)
}
