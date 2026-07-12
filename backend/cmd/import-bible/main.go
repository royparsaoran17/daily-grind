// Command import-bible loads a complete Bible translation into the database
// from the free, no-key API at https://bible.helloao.org (which aggregates
// openly-licensed translations from eBible.org).
//
// Usage:
//
//	go run ./cmd/import-bible                 # defaults: ind_ayt into text_id
//	go run ./cmd/import-bible -translation ind_ayt -lang id
//	go run ./cmd/import-bible -translation eng_web -lang en
//
// It is idempotent: books and verses are upserted, and the curated `meaning`
// notes on existing verses are preserved (only the text column is written).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/royparsaoran/daily-grind/backend/internal/config"
	"github.com/royparsaoran/daily-grind/backend/internal/db"
)

type apiBook struct {
	ID               string `json:"id"`
	CommonName       string `json:"commonName"`
	Name             string `json:"name"`
	NumberOfChapters int    `json:"numberOfChapters"`
	Order            int    `json:"order"`
}

type apiChapter struct {
	Chapter struct {
		Content []struct {
			Type    string        `json:"type"`
			Number  int           `json:"number"`
			Content []interface{} `json:"content"`
		} `json:"content"`
	} `json:"chapter"`
}

func main() {
	translation := flag.String("translation", "ind_ayt", "helloao translation id (e.g. ind_ayt, eng_web)")
	lang := flag.String("lang", "id", "which column to fill: id -> text_id, en -> text_en")
	base := flag.String("base", "https://bible.helloao.org/api", "API base URL")
	workers := flag.Int("workers", 8, "concurrent chapter fetches")
	flag.Parse()

	col := "text_id"
	if *lang == "en" {
		col = "text_en"
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Books
	var booksResp struct {
		Books []apiBook `json:"books"`
	}
	if err := getJSON(client, fmt.Sprintf("%s/%s/books.json", *base, *translation), &booksResp); err != nil {
		log.Fatalf("fetch books: %v", err)
	}
	if len(booksResp.Books) == 0 {
		log.Fatalf("no books returned for translation %q", *translation)
	}
	log.Printf("importing %q (%d books) into %s", *translation, len(booksResp.Books), col)

	for i, b := range booksResp.Books {
		ordinal := b.Order
		if ordinal == 0 {
			ordinal = i + 1
		}
		testament := "OT"
		if ordinal >= 40 { // Matthew onward
			testament = "NT"
		}
		name := b.CommonName
		if name == "" {
			name = b.Name
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO bible_books(id,name,ordinal,chapters,testament)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, ordinal=EXCLUDED.ordinal,
				chapters=EXCLUDED.chapters, testament=EXCLUDED.testament`,
			b.ID, name, ordinal, b.NumberOfChapters, testament); err != nil {
			log.Fatalf("upsert book %s: %v", b.ID, err)
		}
	}

	// 2. Chapters (fan out)
	type job struct {
		bookID  string
		chapter int
	}
	jobs := make(chan job)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalVerses int
	var failures int

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				n, err := importChapter(ctx, client, pool, *base, *translation, col, j.bookID, j.chapter)
				mu.Lock()
				if err != nil {
					failures++
					log.Printf("  ! %s %d: %v", j.bookID, j.chapter, err)
				} else {
					totalVerses += n
				}
				mu.Unlock()
			}
		}()
	}

	for _, b := range booksResp.Books {
		for c := 1; c <= b.NumberOfChapters; c++ {
			jobs <- job{b.ID, c}
		}
	}
	close(jobs)
	wg.Wait()

	log.Printf("done: %d verses imported, %d chapter failures", totalVerses, failures)
	if failures > 0 {
		os.Exit(1)
	}
}

func importChapter(ctx context.Context, client *http.Client, pool *pgxpool.Pool, base, translation, col, bookID string, chapter int) (int, error) {
	var ch apiChapter
	url := fmt.Sprintf("%s/%s/%s/%d.json", base, translation, bookID, chapter)
	if err := getJSON(client, url, &ch); err != nil {
		return 0, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	count := 0
	for _, item := range ch.Chapter.Content {
		if item.Type != "verse" || item.Number == 0 {
			continue
		}
		text := flattenVerse(item.Content)
		if text == "" {
			continue
		}
		// Upsert text into the chosen column without touching `meaning`.
		q := `INSERT INTO bible_verses(book_id,chapter,verse,` + col + `)
		      VALUES ($1,$2,$3,$4)
		      ON CONFLICT (book_id,chapter,verse) DO UPDATE SET ` + col + `=EXCLUDED.` + col
		if _, err := tx.Exec(ctx, q, bookID, chapter, item.Number, text); err != nil {
			return 0, err
		}
		count++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return count, nil
}

// flattenVerse joins the mixed string/object content parts of a verse into
// plain text, skipping footnotes and formatting markers.
func flattenVerse(parts []interface{}) string {
	var sb []byte
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			sb = appendWithSpace(sb, v)
		case map[string]interface{}:
			if t, ok := v["text"].(string); ok {
				sb = appendWithSpace(sb, t)
			}
			// {noteId:...}, {lineBreak:true}, {heading:...} are ignored.
		}
	}
	return string(sb)
}

func appendWithSpace(dst []byte, s string) []byte {
	if s == "" {
		return dst
	}
	if len(dst) > 0 && dst[len(dst)-1] != ' ' {
		dst = append(dst, ' ')
	}
	return append(dst, s...)
}

func getJSON(client *http.Client, url string, out any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
