-- +goose Up
-- +goose StatementBegin
CREATE TABLE pvzs (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     city VARCHAR(50) NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
     date_registered TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pvzs;
-- +goose StatementEnd
