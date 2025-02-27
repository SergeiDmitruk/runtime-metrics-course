-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
