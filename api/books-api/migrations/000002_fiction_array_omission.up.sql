ALTER TABLE fictions
  ADD COLUMN genres_were_omitted boolean NOT NULL DEFAULT false,
  ADD COLUMN tags_were_omitted boolean NOT NULL DEFAULT false;
