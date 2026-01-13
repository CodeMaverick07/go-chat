-- +goose Up
-- +goose StatementBegin
CREATE TABLE tokens (
   hash BYTEA PRIMARY KEY,
   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   expiry TIMESTAMPTZ NOT NULL,
   scope TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tokens;
-- +goose StatementEnd