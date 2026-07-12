package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Fixed IDs keep seeded relations stable and easy to reason about.
const (
	uNadia = "11111111-1111-1111-1111-111111111111"
	uAndi  = "22222222-2222-2222-2222-222222222222"
	uSari  = "33333333-3333-3333-3333-333333333333"
	uBudi  = "44444444-4444-4444-4444-444444444444"
	uRina  = "55555555-5555-5555-5555-555555555555"
)

// seedDemo populates the database with the demo characters and content shown
// in the design. Every seeded account uses the password "password123".
func seedDemo(ctx context.Context, pool *pgxpool.Pool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}
	pw := string(hash)

	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		// Categories -> attribute mapping
		cats := [][4]string{
			{"olahraga", "Olahraga", "ph-barbell", "str"},
			{"kesehatan", "Kesehatan", "ph-heartbeat", "vit"},
			{"belajar", "Belajar", "ph-brain", "int"},
			{"kerja", "Kerja", "ph-briefcase", "wis"},
			{"rohani", "Rohani", "ph-hands-praying", "faith"},
		}
		for _, c := range cats {
			if _, err := tx.Exec(ctx,
				`INSERT INTO categories(id,label,icon,attribute) VALUES ($1,$2,$3,$4)
				 ON CONFLICT (id) DO NOTHING`, c[0], c[1], c[2], c[3]); err != nil {
				return err
			}
		}

		// Users: id, name, email, title, level, exp, coins, streak, str, vit, int, wis, faith
		users := []struct {
			id, name, email, title                string
			level, exp, coins, streak             int
			str, vit, intv, wis, faith            int
		}{
			{uNadia, "Nadia Pramesti", "nadia@email.com", "Petualang Disiplin", 12, 340, 1240, 7, 24, 18, 15, 21, 12},
			{uAndi, "Andi Kurnia", "andi@email.com", "Pelari Tangguh", 15, 210, 1900, 30, 30, 26, 14, 12, 9},
			{uSari, "Sari Wulandari", "sari@email.com", "Sang Konsisten", 13, 120, 1420, 21, 20, 22, 18, 15, 11},
			{uBudi, "Budi Santoso", "budi@email.com", "Ahli Rutinitas", 11, 80, 980, 30, 18, 16, 20, 14, 8},
			{uRina, "Rina Melati", "rina@email.com", "Penjelajah Baru", 9, 200, 620, 4, 12, 14, 16, 10, 15},
		}
		for _, u := range users {
			if _, err := tx.Exec(ctx,
				`INSERT INTO users(id,name,email,password_hash,title,level,exp,coins,streak,str,vit,int_,wis,faith,last_active_on)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,current_date)`,
				u.id, u.name, u.email, pw, u.title, u.level, u.exp, u.coins, u.streak,
				u.str, u.vit, u.intv, u.wis, u.faith); err != nil {
				return err
			}
		}

		// Nadia's quests. doneToday quests are completed today (last_completed_on
		// = today); the rest were done yesterday so their streaks stay alive.
		quests := []struct {
			name, cat, freq, diff string
			exp, coins, streak    int
			reminder              string
			doneToday             bool
		}{
			{"Olahraga 30 menit", "olahraga", "daily", "medium", 50, 15, 12, "Setiap hari, 05.30", true},
			{"Belajar 1 bab", "belajar", "daily", "medium", 40, 12, 5, "Setiap hari, 20.00", true},
			{"Baca Alkitab", "rohani", "daily", "easy", 30, 10, 7, "Setiap hari, 06.00", false},
			{"Minum air 2 liter", "kesehatan", "daily", "easy", 20, 8, 3, "", false},
		}
		for _, q := range quests {
			// last_completed_on: today if done today, else yesterday.
			offset := "current_date - 1"
			if q.doneToday {
				offset = "current_date"
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO quests(user_id,name,category_id,frequency,difficulty,exp_reward,coin_reward,streak,reminder,last_completed_on)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),`+offset+`)`,
				uNadia, q.name, q.cat, q.freq, q.diff, q.exp, q.coins, q.streak, q.reminder); err != nil {
				return err
			}
		}
		// Mark the doneToday quests complete for today.
		if _, err := tx.Exec(ctx, `
			INSERT INTO quest_completions(quest_id,user_id,exp_awarded,coin_awarded)
			SELECT id, user_id, exp_reward, coin_reward FROM quests
			WHERE user_id=$1 AND name IN ('Olahraga 30 menit','Belajar 1 bab')`, uNadia); err != nil {
			return err
		}

		// Friendships (Nadia <-> everyone)
		for _, fid := range []string{uAndi, uSari, uBudi, uRina} {
			if _, err := tx.Exec(ctx,
				`INSERT INTO friendships(user_id,friend_id) VALUES ($1,$2),($2,$1)
				 ON CONFLICT DO NOTHING`, uNadia, fid); err != nil {
				return err
			}
		}

		// Feed posts
		var andiPost string
		if err := tx.QueryRow(ctx,
			`INSERT INTO posts(user_id,body,photo_url,badge)
			 VALUES ($1,$2,$3,$4) RETURNING id`,
			uAndi, "Akhirnya tembus Level 15 setelah 30 hari lari rutin! Semangat terus semua 💪",
			"https://images.unsplash.com/photo-1476480862126-209bfaa8edc8?w=800", "Naik Lvl 15").Scan(&andiPost); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO posts(user_id,body,badge)
			 VALUES ($1,$2,$3)`,
			uSari, "Selesai lari 5K pagi ini, VIT naik lagi! Cuaca lagi enak buat olahraga ✨", "+50 EXP"); err != nil {
			return err
		}
		// Likes + comments on Andi's post
		for _, liker := range []string{uNadia, uSari, uBudi} {
			if _, err := tx.Exec(ctx, `INSERT INTO post_likes(post_id,user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, andiPost, liker); err != nil {
				return err
			}
		}
		comments := []struct{ uid, body string }{
			{uSari, "Keren banget, inspiratif! 🔥"},
			{uBudi, "Gas bareng besok pagi 🏃"},
		}
		for _, c := range comments {
			if _, err := tx.Exec(ctx, `INSERT INTO comments(post_id,user_id,body) VALUES ($1,$2,$3)`, andiPost, c.uid, c.body); err != nil {
				return err
			}
		}

		// Bible passages + devotional content pool (see content_seed.go).
		if err := seedBible(ctx, tx); err != nil {
			return err
		}
		if err := seedDevotionalPool(ctx, tx); err != nil {
			return err
		}
		return nil
	})
}
