-- +goose Up
  DROP INDEX IF EXISTS idx_unique_orig_url;
  CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_orig_url ON public.urls (original_url,user_id);
-- +goose StatementBegin

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_orig_url ON public.urls (original_url);
-- +goose StatementEnd
