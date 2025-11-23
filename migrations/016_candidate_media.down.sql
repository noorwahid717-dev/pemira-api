-- +goose Down

ALTER TABLE candidates
    DROP COLUMN IF EXISTS updated_by_admin_id,
    DROP COLUMN IF EXISTS photo_media_id;

DROP TABLE IF EXISTS candidate_media;
