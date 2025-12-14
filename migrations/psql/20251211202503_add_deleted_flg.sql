-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.urls
    ADD COLUMN is_deleted boolean NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.urls DROP COLUMN IF EXISTS is_deleted;
-- +goose StatementEnd
