ALTER TABLE chat ADD COLUMN blog BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE message ADD COLUMN blog_post BOOLEAN NOT NULL DEFAULT FALSE;
