-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.urls
    ADD COLUMN user_id character varying NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.urls DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
