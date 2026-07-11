# DailyGrind — RPG Habit Tracker

A mobile web app that turns daily routines into a light RPG: complete **quests**,
earn **EXP** and **coins**, level up, and grow five character attributes
(**STR / VIT / INT / WIS / FAITH**) mapped from habit categories. Includes a social
feed & leaderboard, a Bible reader, and a daily devotional. UI is in Bahasa
Indonesia with light + dark themes.

Implemented from the Claude Design source `DailyGrind App.dc.html`
(see [design/DESIGN_SPEC.md](design/DESIGN_SPEC.md)).

| Layer    | Stack |
|----------|-------|
| Frontend | Nuxt 3 (Vue 3, Pinia) — mobile-first web app |
| Backend  | Go 1.24 (stdlib `net/http`, pgx) + JWT auth |
| Database | PostgreSQL 16 |

## Screens

1. **Auth** — login / register (demo: `nadia@email.com` / `password123`)
2. **Beranda** (Home) — level card, attributes, today's quests, friend activity
3. **Quest** — daily/weekly/monthly quests with completion + streaks
4. **Buat Quest** — create a quest; category → attribute, difficulty → reward
5. **Profil** — character sheet: EXP, attribute bars, achievements
6. **Teman** — friends + weekly EXP leaderboard
7. **Alkitab** (Bible) — full book/chapter picker, verses with "meaning" cards (EN/ID)
8. **Aktivitas** (Feed) — posts, likes, replies
9. **Renungan Harian** — daily devotional; completing grants FAITH
10. **Jurnal** — daily journaling with mood, editable history
11. **Analisis** — progress charts: 14-day activity, by-category, headline stats

## Quick start

### 1. Database + backend (Docker)

```bash
# from the repo root
docker compose up -d db          # Postgres on :5432
cd backend
cp .env.example .env
go run ./cmd/api                 # migrates + seeds on first boot, serves :8080
```

Or run the whole backend in Docker:

```bash
docker compose up --build        # db + backend
```

### 2. Frontend

> **Requires Node 20+** (the Nuxt CLI uses `node:util.styleText`). A `.nvmrc`
> pins Node 22 — run `nvm use` in `frontend/`.

```bash
cd frontend
nvm use                          # -> Node 22
npm install
npm run dev                      # http://localhost:3000
```

Open http://localhost:3000 and log in with the demo account.

## Configuration

**Backend** (`backend/.env`):

| Var | Default |
|-----|---------|
| `PORT` | `8080` |
| `DATABASE_URL` | `postgres://dailygrind:dailygrind@localhost:5432/dailygrind?sslmode=disable` |
| `JWT_SECRET` | dev secret — **change in production** |
| `CORS_ORIGIN` | `http://localhost:3000` |

**Frontend** (`frontend/.env`): `NUXT_PUBLIC_API_BASE` (default `http://localhost:8080/api`).

## API

All routes are prefixed with `/api`. Authenticated routes require
`Authorization: Bearer <token>`. **Wire format is `snake_case`** to match the
database convention (e.g. `next_exp`, `exp_reward`, `day_of_month`); the Nuxt
client transparently converts to/from camelCase in `composables/useApi.ts`.

Quests support scheduling: `weekly` quests carry a `weekday` (0=Sunday..6),
`monthly` quests a `day_of_month` (1..31); completion is tracked **per period**
(day/week/month), so a weekly quest counts as done once for the whole week.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Create account → `{token, user}` |
| POST | `/auth/login` | Log in → `{token, user}` |
| GET | `/me` | Current character |
| PUT | `/me` | Update profile (name, title) |
| GET | `/achievements` | Badges with live unlock state + progress |
| GET | `/analytics` | Progress/history: 14-day series, by-category, totals |
| GET | `/categories` | Habit categories + attribute mapping |
| GET | `/quests` | List quests with today's completion state |
| POST | `/quests` | Create a quest |
| POST | `/quests/{id}/complete` | Complete today (awards EXP/coins/attribute; advances streaks) |
| DELETE | `/quests/{id}/complete` | Undo today's completion |
| DELETE | `/quests/{id}` | Archive a quest |
| POST | `/streak/freeze` | Buy a streak freeze (100 coins) |
| GET | `/friends` | Friends + weekly-EXP leaderboard |
| GET | `/users/search?q=` | Search users to add |
| POST | `/friends/{id}` | Add a friend |
| DELETE | `/friends/{id}` | Remove a friend |
| GET | `/feed` | Activity feed with likes & comments |
| POST | `/feed` | Create a post |
| POST | `/feed/{id}/like` | Toggle like |
| POST | `/feed/{id}/comments` | Add a comment |
| GET | `/bible?bookId=PSA&chapter=23` | Verses (also accepts `?book=Mazmur`; defaults to Mazmur 23) |
| GET | `/bible/books` | Canonical books + which chapters have verses (for the picker) |
| GET | `/devotional/today` | Today's devotional (materialized from the pool) |
| POST | `/devotional/{id}/complete` | Complete devotional (+FAITH) |
| GET | `/journal` | List all journal entries (newest first) |
| GET | `/journal/{date}` | Get entry for a date (`today` or `YYYY-MM-DD`) |
| PUT | `/journal/{date}` | Create/update the entry for a date |
| DELETE | `/journal/{date}` | Delete the entry for a date |

## Game rules

- **Leveling**: EXP to reach the next level = `200 + level*25` (level 12 → 13 costs 500, per the design).
- **Rewards by difficulty**: easy `+20 EXP / +8`, medium `+40 / +15`, hard `+60 / +25`.
- **Attributes** grow +1 each time a quest of the mapped category is completed:
  Olahraga→STR, Kesehatan→VIT, Belajar→INT, Kerja→WIS, Rohani→FAITH.
- Completions are idempotent per calendar day; undo reverses EXP/coins/attribute
  (with level roll-back).

### Streaks & freezes

- **Per-quest streak** counts consecutive completions within the frequency
  window (daily = 1 day, weekly = 7, monthly = 31). Miss the window and the
  streak shows `0` — the app reads streaks break-aware, no cron required.
- **Daily-activity streak** (the fire chip) advances when you complete any quest
  or devotional. Missed days are bridged by **streak freezes** if you have
  enough banked; otherwise the streak resets. Freezes are consumed at your next
  activity, and are bought for 100 coins (`POST /streak/freeze`). New users start
  with 2.

## Content: Bible & devotionals

**Bible** — the app ships with a few curated passages (with "meaning" notes) in
[`backend/internal/db/content_seed.go`](backend/internal/db/content_seed.go) so
the reader works with zero setup. To load the **complete Bible**, run the
importer:

```bash
cd backend
# Full Indonesian "Alkitab Yang Terbuka" (66 books, 1,189 chapters, ~31k verses)
go run ./cmd/import-bible                          # defaults: -translation ind_ayt -lang id
# Optionally add an English column too:
go run ./cmd/import-bible -translation eng_web -lang en
```

The importer pulls from the free, no-key [helloao Bible API](https://bible.helloao.org)
(openly-licensed translations aggregated from eBible.org). It is **idempotent**
and **text-only-upsert**, so re-running is safe and it never clobbers the curated
`meaning` notes — run the demo seed first, then the importer. Data lives in the
`dg_pgdata` Docker volume and survives restarts; you only re-import after
`docker compose down -v`. Books/chapters keyed by USFM code (`GEN`, `PSA`, `JHN`),
so curated and imported content never diverge.

**Journaling** — a daily journal (one editable entry per calendar day) with
title, free-text body, an optional mood, and an optional `prompt` (e.g. the
devotional passage that inspired it). Reachable from the **Alkitab** and
**Renungan** screens; the devotional's "Tulis jurnal" link pre-fills the prompt.
Entries list newest-first and are individually editable/deletable.

**Daily devotionals** — instead of pre-writing every date, a **content pool**
(`devotionalPool`) is seeded once and one entry is *materialized* into the
`devotionals` table per calendar day, chosen deterministically from the date
(`day_number mod pool_size`). So:

- there is always exactly one devotional for "today", offline, with zero
  maintenance;
- everyone sees the same one each day and it's stable across requests;
- completing it grants FAITH + EXP and counts toward the daily streak.

To scale beyond the pool, replace `materializeToday` with a scheduled job that
writes an editorially- or LLM-authored devotional for the next day into the same
`devotionals` table — the API and UI need no changes.

## Project layout

```
daily-grind/
├─ backend/            Go API
│  ├─ cmd/api/         main entrypoint
│  ├─ cmd/import-bible/ full-Bible importer (helloao API -> Postgres)
│  └─ internal/
│     ├─ api/          handlers, router, middleware
│     ├─ auth/         JWT + bcrypt
│     ├─ config/       env loading
│     ├─ db/           pgx pool, schema.sql, Go seeder
│     └─ models/       JSON DTOs
├─ frontend/           Nuxt 3 app
│  ├─ pages/           one file per screen
│  ├─ components/      TabBar, QuestRow, ThemeToggle, …
│  ├─ composables/     useApi, useQuests, useTheme, useToast
│  ├─ stores/          Pinia auth store
│  └─ assets/css/      design tokens + component styles
├─ design/             original design + spec
└─ docker-compose.yml
```
