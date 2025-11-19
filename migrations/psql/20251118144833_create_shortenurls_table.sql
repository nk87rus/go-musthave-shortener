-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.urls
(
    uuid character varying NOT NULL,
    short_url text NOT NULL,
    orig_url text NOT NULL,
    PRIMARY KEY (uuid)
);
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.urls;
-- +goose StatementEnd
