-- 002: Rename email → username if old schema exists (safe: skips if already renamed)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'email'
    ) THEN
        ALTER TABLE users RENAME COLUMN email TO username;
    END IF;
END $$;

-- 002b: Add phone column if missing (old schema may not have it)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'phone'
    ) THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(50);
    END IF;
END $$;
