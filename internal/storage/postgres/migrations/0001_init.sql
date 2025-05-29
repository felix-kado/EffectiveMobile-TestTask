-- internal/storage/migrations/0001_init.sql

-- +goose Up
CREATE TABLE persons (
                         id SERIAL PRIMARY KEY,
                         name VARCHAR(100) NOT NULL,
                         surname VARCHAR(100) NOT NULL,
                         patronymic VARCHAR(100),
                         age INT,
                         gender VARCHAR(10),
                         nationality VARCHAR(100),
                         created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                         updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS persons;
