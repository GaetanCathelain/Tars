-- Make github_id nullable so email-only users can exist.
ALTER TABLE users ALTER COLUMN github_id DROP NOT NULL;

-- Add password_hash for email/password auth users.
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- Enforce unique emails for email-based login lookup.
-- email was already present but had no uniqueness constraint.
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
