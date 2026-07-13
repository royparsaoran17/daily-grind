-- DailyGrind schema (idempotent). Applied on every boot.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users / characters ---------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    title         TEXT NOT NULL DEFAULT 'Petualang Baru',
    level         INT  NOT NULL DEFAULT 1,
    exp           INT  NOT NULL DEFAULT 0,
    coins         INT  NOT NULL DEFAULT 0,
    streak        INT  NOT NULL DEFAULT 0,
    -- RPG attributes
    str   INT NOT NULL DEFAULT 0,
    vit   INT NOT NULL DEFAULT 0,
    int_  INT NOT NULL DEFAULT 0,
    wis   INT NOT NULL DEFAULT 0,
    faith INT NOT NULL DEFAULT 0,
    -- Daily-activity streak bookkeeping
    streak_freezes INT NOT NULL DEFAULT 2,   -- coin-purchasable "grace days"
    last_active_on DATE,                     -- last day that counted toward the streak
    onboarded_at   TIMESTAMPTZ,              -- null until the user finishes onboarding
    locale         TEXT NOT NULL DEFAULT 'id' CHECK (locale IN ('id','en')),
    avatar_url     TEXT,
    timezone       TEXT NOT NULL DEFAULT 'Asia/Jakarta',  -- IANA tz for date math
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Categories map a habit type to an attribute --------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id        TEXT PRIMARY KEY,          -- slug e.g. 'olahraga'
    label     TEXT NOT NULL,
    icon      TEXT NOT NULL,             -- phosphor icon name
    attribute TEXT NOT NULL CHECK (attribute IN ('str','vit','int','wis','faith'))
);

-- Quests / habits ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS quests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    category_id  TEXT NOT NULL REFERENCES categories(id),
    frequency    TEXT NOT NULL DEFAULT 'daily' CHECK (frequency IN ('daily','weekly','monthly')),
    difficulty   TEXT NOT NULL DEFAULT 'medium' CHECK (difficulty IN ('easy','medium','hard')),
    exp_reward   INT NOT NULL DEFAULT 30,
    coin_reward  INT NOT NULL DEFAULT 10,
    reminder     TEXT,
    weekday      INT,                        -- weekly quests: 0=Sunday..6=Saturday
    day_of_month INT,                        -- monthly quests: 1..31
    streak       INT NOT NULL DEFAULT 0,
    last_completed_on DATE,                  -- for streak-break detection
    archived     BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_quests_user ON quests(user_id);

-- One completion per quest per calendar day ----------------------------------
CREATE TABLE IF NOT EXISTS quest_completions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quest_id     UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed_on DATE NOT NULL DEFAULT current_date,
    exp_awarded  INT NOT NULL,
    coin_awarded INT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (quest_id, completed_on)
);
CREATE INDEX IF NOT EXISTS idx_completions_user ON quest_completions(user_id);

-- Friendships (symmetric, stored one row per direction for simple queries) ----
CREATE TABLE IF NOT EXISTS friendships (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id)
);

-- Pending friend requests (accepted ones become rows in friendships) ---------
CREATE TABLE IF NOT EXISTS friend_requests (
    from_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (from_id, to_id)
);
CREATE INDEX IF NOT EXISTS idx_freq_to ON friend_requests(to_id);

-- Social feed ----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    photo_url  TEXT,
    badge      TEXT,                      -- optional tag e.g. 'Naik Lvl 15'
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);

CREATE TABLE IF NOT EXISTS post_likes (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE IF NOT EXISTS comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);

-- Bible ----------------------------------------------------------------------
-- Canonical book list (66 books). Populated by the seed (a few books) and, in
-- full, by `cmd/import-bible`.
CREATE TABLE IF NOT EXISTS bible_books (
    id        TEXT PRIMARY KEY,       -- USFM code: GEN, PSA, JHN, …
    name      TEXT NOT NULL,          -- display name (Bahasa Indonesia)
    ordinal   INT  NOT NULL,          -- canonical order 1..66
    chapters  INT  NOT NULL,          -- total chapters in the book
    testament TEXT NOT NULL DEFAULT 'OT' CHECK (testament IN ('OT','NT'))
);

CREATE TABLE IF NOT EXISTS bible_verses (
    book_id  TEXT NOT NULL REFERENCES bible_books(id) ON DELETE CASCADE,
    chapter  INT  NOT NULL,
    verse    INT  NOT NULL,
    text_id  TEXT NOT NULL,          -- Bahasa Indonesia
    text_en  TEXT,                   -- English (optional second import)
    meaning  TEXT,                   -- optional curated explanation
    PRIMARY KEY (book_id, chapter, verse)
);

-- Daily devotionals ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS devotionals (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    for_date   DATE NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    passage    TEXT NOT NULL,        -- e.g. 'Mazmur 23:1-3'
    verse_text TEXT NOT NULL,
    reflection TEXT NOT NULL,
    prayer     TEXT NOT NULL,
    faith_reward INT NOT NULL DEFAULT 20
);

CREATE TABLE IF NOT EXISTS devotional_completions (
    devotional_id UUID NOT NULL REFERENCES devotionals(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (devotional_id, user_id)
);

-- Reusable devotional content. Each day's devotional is materialized into
-- `devotionals` by picking a pool entry deterministically from the date, so
-- there is always exactly one for today without pre-populating every date.
CREATE TABLE IF NOT EXISTS devotional_pool (
    id         INT PRIMARY KEY,     -- stable ordering index (0-based)
    title      TEXT NOT NULL,
    passage    TEXT NOT NULL,
    verse_text TEXT NOT NULL,
    reflection TEXT NOT NULL,
    prayer     TEXT NOT NULL,
    faith_reward INT NOT NULL DEFAULT 20
);

-- Bible highlights & bookmarks ----------------------------------------------
CREATE TABLE IF NOT EXISTS bible_marks (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id    TEXT NOT NULL,
    chapter    INT  NOT NULL,
    verse      INT  NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('highlight','bookmark')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, book_id, chapter, verse, kind)
);
CREATE INDEX IF NOT EXISTS idx_marks_user ON bible_marks(user_id, kind);

-- Bible reading plans (plan CONTENT is defined in code; DB tracks progress) ---
CREATE TABLE IF NOT EXISTS reading_plan_enrollments (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id    TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, plan_id)
);
CREATE TABLE IF NOT EXISTS reading_plan_progress (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id      TEXT NOT NULL,
    day_no       INT  NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, plan_id, day_no)
);

-- Prayer list ----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS prayers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    answered    BOOLEAN NOT NULL DEFAULT false,
    answered_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_prayers_user ON prayers(user_id, answered, created_at DESC);

-- Journaling: one editable entry per user per calendar day ------------------
CREATE TABLE IF NOT EXISTS journal_entries (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_date DATE NOT NULL,
    title      TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    mood       TEXT,                    -- optional: senang/biasa/lelah/sedih/bersyukur
    prompt     TEXT,                    -- optional context, e.g. a devotional passage
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, entry_date)
);
CREATE INDEX IF NOT EXISTS idx_journal_user_date ON journal_entries(user_id, entry_date DESC);

-- Idempotent migrations for databases created before these columns existed.
ALTER TABLE users  ADD COLUMN IF NOT EXISTS streak_freezes INT NOT NULL DEFAULT 2;
ALTER TABLE users  ADD COLUMN IF NOT EXISTS last_active_on DATE;
ALTER TABLE quests ADD COLUMN IF NOT EXISTS last_completed_on DATE;
ALTER TABLE quests ADD COLUMN IF NOT EXISTS weekday INT;
ALTER TABLE quests ADD COLUMN IF NOT EXISTS day_of_month INT;
ALTER TABLE users  ADD COLUMN IF NOT EXISTS onboarded_at TIMESTAMPTZ;
ALTER TABLE users  ADD COLUMN IF NOT EXISTS locale TEXT NOT NULL DEFAULT 'id';
ALTER TABLE users  ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE users  ADD COLUMN IF NOT EXISTS timezone TEXT NOT NULL DEFAULT 'Asia/Jakarta';
