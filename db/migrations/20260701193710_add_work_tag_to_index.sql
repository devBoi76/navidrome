-- +goose Up
ALTER TABLE media_file ADD COLUMN work varchar(255) DEFAULT '' NOT NULL;

-- +goose Down
ALTER TABLE media_file DROP COLUMN work;
