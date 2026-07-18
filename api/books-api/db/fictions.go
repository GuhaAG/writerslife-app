package db

import (
	"context"
	"fmt"

	"github.com/guhaag/writerslife-books-api/model"
)

const fictionColumns = `id, author_id, author_name, title, synopsis, cover_url, genres, tags, genres_were_omitted, tags_were_omitted, status, created_at, updated_at, follower_count, view_count, chapter_count`

func (s *Store) CreateFiction(ctx context.Context, fiction *model.Fiction) error {
	fiction.GenresWereOmitted = fiction.Genres == nil
	fiction.TagsWereOmitted = fiction.Tags == nil
	err := s.pool.QueryRow(ctx, `
		INSERT INTO fictions (id, author_id, author_name, title, synopsis, cover_url, genres, tags, genres_were_omitted, tags_were_omitted, status, created_at, updated_at)
		VALUES ('f' || nextval('fiction_id_sequence'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+fictionColumns,
		fiction.AuthorID, fiction.AuthorName, fiction.Title, fiction.Synopsis, fiction.CoverURL, nonNilStrings(fiction.Genres), nonNilStrings(fiction.Tags), fiction.GenresWereOmitted, fiction.TagsWereOmitted, fiction.Status, fiction.CreatedAt, fiction.UpdatedAt,
	).Scan(fictionFields(fiction)...)
	if err != nil {
		return err
	}
	normalizeFictionArrays(fiction)
	return nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (s *Store) GetFiction(ctx context.Context, id string) (*model.Fiction, error) {
	return s.findFiction(ctx, `SELECT `+fictionColumns+` FROM fictions WHERE id = $1`, id)
}

func (s *Store) ListFictions(ctx context.Context, search, genre, sort string) ([]*model.Fiction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+fictionColumns+` FROM fictions
		WHERE ($1 = '' OR strpos(lower(title), lower($1)) > 0 OR strpos(lower(synopsis), lower($1)) > 0)
		  AND ($2 = '' OR EXISTS (SELECT 1 FROM unnest(genres) AS item WHERE lower(item) = lower($2)))
		ORDER BY CASE WHEN $3 = 'popular' THEN view_count END DESC,
		         CASE WHEN $3 <> 'popular' THEN updated_at END DESC`, search, genre, sort)
	if err != nil {
		return nil, fmt.Errorf("list fictions: %w", err)
	}
	defer rows.Close()

	return scanFictions(rows)
}

func (s *Store) ListFictionsByAuthor(ctx context.Context, authorID string) ([]*model.Fiction, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+fictionColumns+` FROM fictions WHERE author_id = $1 ORDER BY updated_at DESC`, authorID)
	if err != nil {
		return nil, fmt.Errorf("list author fictions: %w", err)
	}
	defer rows.Close()
	return scanFictions(rows)
}

func (s *Store) UpdateFiction(ctx context.Context, fiction *model.Fiction) error {
	err := s.pool.QueryRow(ctx, `
		UPDATE fictions SET title = $1, synopsis = $2, cover_url = $3, genres = $4, tags = $5, genres_were_omitted = $6, tags_were_omitted = $7, status = $8, updated_at = $9 WHERE id = $10
		RETURNING `+fictionColumns,
		fiction.Title, fiction.Synopsis, fiction.CoverURL, nonNilStrings(fiction.Genres), nonNilStrings(fiction.Tags), fiction.GenresWereOmitted, fiction.TagsWereOmitted, fiction.Status, fiction.UpdatedAt, fiction.ID,
	).Scan(fictionFields(fiction)...)
	if err != nil {
		return err
	}
	normalizeFictionArrays(fiction)
	return nil
}

func (s *Store) IncrementFictionViews(ctx context.Context, id string) (*model.Fiction, error) {
	return s.findFiction(ctx, `UPDATE fictions SET view_count = view_count + 1 WHERE id = $1 RETURNING `+fictionColumns, id)
}

type rowScanner interface {
	Scan(...interface{}) error
}

func (s *Store) findFiction(ctx context.Context, query, id string) (*model.Fiction, error) {
	fiction := &model.Fiction{}
	if err := s.pool.QueryRow(ctx, query, id).Scan(fictionFields(fiction)...); err != nil {
		return nil, fmt.Errorf("find fiction: %w", err)
	}
	normalizeFictionArrays(fiction)
	return fiction, nil
}

func scanFictions(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]*model.Fiction, error) {
	result := make([]*model.Fiction, 0)
	for rows.Next() {
		fiction := &model.Fiction{}
		if err := rows.Scan(fictionFields(fiction)...); err != nil {
			return nil, fmt.Errorf("scan fiction: %w", err)
		}
		normalizeFictionArrays(fiction)
		result = append(result, fiction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fictions: %w", err)
	}
	return result, nil
}

func fictionFields(fiction *model.Fiction) []interface{} {
	return []interface{}{&fiction.ID, &fiction.AuthorID, &fiction.AuthorName, &fiction.Title, &fiction.Synopsis, &fiction.CoverURL,
		&fiction.Genres, &fiction.Tags, &fiction.GenresWereOmitted, &fiction.TagsWereOmitted, &fiction.Status, &fiction.CreatedAt, &fiction.UpdatedAt, &fiction.FollowerCount,
		&fiction.ViewCount, &fiction.ChapterCount}
}

func normalizeFictionArrays(fiction *model.Fiction) {
	if fiction.GenresWereOmitted {
		fiction.Genres = nil
	}
	if fiction.TagsWereOmitted {
		fiction.Tags = nil
	}
}
