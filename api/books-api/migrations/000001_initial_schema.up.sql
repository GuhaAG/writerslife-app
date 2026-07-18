CREATE TABLE users (
  id text PRIMARY KEY,
  username text NOT NULL UNIQUE,
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  bio text NOT NULL DEFAULT '',
  avatar_url text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL
);

CREATE TABLE fictions (
  id text PRIMARY KEY,
  author_id text NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  author_name text NOT NULL,
  title text NOT NULL,
  synopsis text NOT NULL DEFAULT '',
  cover_url text NOT NULL DEFAULT '',
  genres text[] NOT NULL DEFAULT '{}',
  tags text[] NOT NULL DEFAULT '{}',
  status text NOT NULL DEFAULT 'ongoing' CHECK (status IN ('ongoing', 'completed', 'hiatus')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  follower_count integer NOT NULL DEFAULT 0 CHECK (follower_count >= 0),
  view_count integer NOT NULL DEFAULT 0 CHECK (view_count >= 0),
  chapter_count integer NOT NULL DEFAULT 0 CHECK (chapter_count >= 0)
);

CREATE TABLE chapters (
  id text PRIMARY KEY,
  fiction_id text NOT NULL REFERENCES fictions(id) ON DELETE CASCADE,
  title text NOT NULL,
  content text NOT NULL DEFAULT '',
  chapter_number integer NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
  published_at timestamptz NULL,
  created_at timestamptz NOT NULL,
  UNIQUE (fiction_id, chapter_number)
);

CREATE TABLE follows (
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  fiction_id text NOT NULL REFERENCES fictions(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, fiction_id)
);

CREATE SEQUENCE user_id_sequence START 3;
CREATE SEQUENCE fiction_id_sequence START 4;
CREATE SEQUENCE chapter_id_sequence START 8;

CREATE INDEX fictions_author_id_idx ON fictions(author_id);
CREATE INDEX fictions_updated_at_idx ON fictions(updated_at DESC);
CREATE INDEX fictions_view_count_idx ON fictions(view_count DESC);
CREATE INDEX chapters_fiction_number_idx ON chapters(fiction_id, chapter_number);
CREATE INDEX follows_user_id_idx ON follows(user_id);
