-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   email TEXT UNIQUE NOT NULL,
   hashed_password TEXT NOT NULL,
   role TEXT CHECK (role IN ('client', 'moderator', 'employee')) NOT NULL,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
