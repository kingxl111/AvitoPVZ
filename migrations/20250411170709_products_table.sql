-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   reception_id UUID NOT NULL,
   added_at TIMESTAMP NOT NULL DEFAULT NOW(),
   product_type VARCHAR(20) NOT NULL CHECK (product_type IN ('электроника', 'одежда', 'обувь')),
   CONSTRAINT fk_receptions
       FOREIGN KEY (reception_id)
           REFERENCES receptions (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
