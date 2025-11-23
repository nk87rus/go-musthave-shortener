-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.urls
(
    uuid character varying NOT NULL,
    short_url text NOT NULL,
    original_url text NOT NULL,
    CONSTRAINT urls_pkey PRIMARY KEY (uuid),
    CONSTRAINT uuid_uniq UNIQUE (uuid)
);
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.urls;
-- +goose StatementEnd
