package db

import (
	"context"
	"fmt"
	"time"

	"github.com/guhaag/writerslife-books-api/model"
	"github.com/jackc/pgx/v5"
)

const chapterColumns = `id, fiction_id, title, content, chapter_number, status, COALESCE(published_at, '0001-01-01 00:00:00+00'::timestamptz), created_at`

func (s *Store) ListChapters(ctx context.Context, fictionID string, includeDrafts bool) ([]*model.Chapter, error) {
	query := `SELECT ` + chapterColumns + ` FROM chapters WHERE fiction_id = $1`
	if !includeDrafts {
		query += ` AND status = 'published'`
	}
	query += ` ORDER BY chapter_number`
	rows, err := s.pool.Query(ctx, query, fictionID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()
	return scanChapters(rows)
}
func (s *Store) GetChapter(ctx context.Context, fictionID string, number int, includeDrafts bool) (*model.Chapter, error) {
	query := `SELECT ` + chapterColumns + ` FROM chapters WHERE fiction_id = $1 AND chapter_number = $2`
	if !includeDrafts {
		query += ` AND status = 'published'`
	}
	chapter := &model.Chapter{}
	err := s.pool.QueryRow(ctx, query, fictionID, number).Scan(chapterFields(chapter)...)
	if err != nil {
		return nil, fmt.Errorf("get chapter: %w", err)
	}
	return chapter, nil
}
func (s *Store) CreateChapter(ctx context.Context, chapter *model.Chapter) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT 1 FROM fictions WHERE id = $1 FOR UPDATE`, chapter.FictionID); err != nil {
		return fmt.Errorf("lock fiction: %w", err)
	}
	var number int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(chapter_number), 0) + 1 FROM chapters WHERE fiction_id = $1`, chapter.FictionID).Scan(&number); err != nil {
		return err
	}
	chapter.ChapterNumber = number
	if chapter.Status == "published" {
		chapter.PublishedAt = chapter.CreatedAt
		if _, err = tx.Exec(ctx, `UPDATE fictions SET chapter_count = chapter_count + 1, updated_at = $2 WHERE id = $1`, chapter.FictionID, chapter.CreatedAt); err != nil {
			return err
		}
	}
	err = tx.QueryRow(ctx, `INSERT INTO chapters (id, fiction_id, title, content, chapter_number, status, published_at, created_at) VALUES ('c' || nextval('chapter_id_sequence'), $1,$2,$3,$4,$5,NULLIF($6, '0001-01-01 00:00:00+00'::timestamptz),$7) RETURNING `+chapterColumns, chapter.FictionID, chapter.Title, chapter.Content, chapter.ChapterNumber, chapter.Status, chapter.PublishedAt, chapter.CreatedAt).Scan(chapterFields(chapter)...)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) UpdateChapter(ctx context.Context, chapter *model.Chapter) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var oldStatus string
	if err = tx.QueryRow(ctx, `SELECT status FROM chapters WHERE fiction_id = $1 AND chapter_number = $2 FOR UPDATE`, chapter.FictionID, chapter.ChapterNumber).Scan(&oldStatus); err != nil {
		return fmt.Errorf("lock chapter: %w", err)
	}
	chapterCountDelta := 0
	if oldStatus == "draft" && chapter.Status == "published" {
		chapter.PublishedAt = time.Now()
		chapterCountDelta = 1
	} else if oldStatus == "published" && chapter.Status == "draft" {
		chapter.PublishedAt = time.Time{}
	}
	if chapterCountDelta != 0 {
		if _, err = tx.Exec(ctx, `UPDATE fictions SET chapter_count = GREATEST(chapter_count + $2, 0), updated_at = $3 WHERE id = $1`, chapter.FictionID, chapterCountDelta, time.Now()); err != nil {
			return err
		}
	}
	err = tx.QueryRow(ctx, `UPDATE chapters SET title=$1, content=$2, status=$3, published_at=CASE WHEN $3='published' THEN NULLIF($4, '0001-01-01 00:00:00+00'::timestamptz) ELSE NULL END WHERE fiction_id=$5 AND chapter_number=$6 RETURNING `+chapterColumns, chapter.Title, chapter.Content, chapter.Status, chapter.PublishedAt, chapter.FictionID, chapter.ChapterNumber).Scan(chapterFields(chapter)...)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func scanChapters(rows pgx.Rows) ([]*model.Chapter, error) {
	result := make([]*model.Chapter, 0)
	for rows.Next() {
		c := &model.Chapter{}
		if err := rows.Scan(chapterFields(c)...); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
func chapterFields(c *model.Chapter) []interface{} {
	return []interface{}{&c.ID, &c.FictionID, &c.Title, &c.Content, &c.ChapterNumber, &c.Status, &c.PublishedAt, &c.CreatedAt}
}
