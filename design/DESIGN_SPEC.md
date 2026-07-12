# DailyGrind — Design Spec

Source: Claude Design project `07a56fc6-0821-48a8-8b22-880b8f40eb6b`, file `DailyGrind App.dc.html`.

Mobile RPG-style habit tracker (Bahasa Indonesia). Turn routines into a light RPG: daily quests, level up, and skills that grow from habit categories. Minimal, warm aesthetic, light + dark mode.

## Design tokens

Fonts: **Space Grotesk** (headings/numbers), **Plus Jakarta Sans** (body). Icons: Phosphor.

Light theme:
- `--primary` maroon `#8c2f3a`, `--primary2` `#a8434e`
- `--bg` `#f7f6f2`, `--surface` `#ffffff`, `--ink` `#1b1a17`, `--muted` `#8a867d`
- amber (EXP/gold) `#f2a63b`, green (done/streak) `#2c9c68`

Dark theme:
- `--primary` `#cf5b67`, `--bg` `#1c1a19`, `--surface` `#262321`, `--ink` `#f1ede7`

Attribute colors: STR `#e0574f`, VIT `#2c9c68`, INT `#3a7bd5`, WIS `#b8434f`, FAITH `#c88a1c`.

## Screens

1. **Auth** — login/register toggle, email + password, Google button, EN/ID switch.
2. **Home (Beranda)** — greeting, streak chip, level/EXP card with avatar ring, coins, 5 attribute mini-stats, today's quests (checkable), friend activity feed.
3. **Quests** — daily/weekly/monthly segmented tabs, progress card, quest list with category→skill mapping, streak per quest, EXP reward, FAB to add.
4. **Create Quest** — name, frequency, category (maps to a skill), difficulty (sets EXP/coin reward), reminder.
5. **Profile / Character** — avatar ring w/ level, streak/coins/quest-count chips, EXP bar, attribute bars, achievement badges.
6. **Friends** — search, weekly leaderboard, friend list with streak + level.
7. **Bible (Alkitab)** — book/chapter selector, verses with highlight, AI "meaning" card, action bar (highlight/save/meaning/share).
8. **Activity Feed** — composer, posts with photo, likes, replies.
9. **Daily Devotional (Renungan Harian)** — hero, verse, reflection, prayer, complete button (+FAITH), streak.

## Domain model

- **User**: name, email, level, exp, coins, streak, joined date, title.
- **Attributes** (per user): STR, VIT, INT, WIS, FAITH values.
- **Category**: name, icon, maps to one attribute (Olahraga→STR, Belajar→INT, Rohani→FAITH, Kesehatan→VIT, Kerja→WIS...).
- **Quest**: name, category, frequency (daily/weekly/monthly), difficulty (easy/medium/hard → exp+coins), reminder, streak.
- **QuestCompletion**: quest, user, date, exp/coins awarded.
- **Friendship**: user ↔ friend.
- **Post**: author, body, photo, likes; **Comment**: post, author, body.
- **BibleVerse**: book, chapter, verse, text (id/en), optional meaning.
- **Devotional**: date, title, passage, verse text, reflection, prayer.
