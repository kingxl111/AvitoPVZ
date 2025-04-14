-- +goose Up
-- +goose StatementBegin
CREATE TABLE receptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pvz_id UUID NOT NULL,
    reception_date TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL CHECK (status IN ('in_progress', 'close')),
    CONSTRAINT fk_pvz
        FOREIGN KEY (pvz_id)
            REFERENCES pvzs (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS receptions;
-- +goose StatementEnd
