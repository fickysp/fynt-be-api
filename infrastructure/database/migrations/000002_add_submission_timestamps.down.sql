ALTER TABLE submissions
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS decided_at;
