package models

import "time"

// Attributes are the five RPG stats grown by habit categories.
type Attributes struct {
	STR   int `json:"str"`
	VIT   int `json:"vit"`
	INT   int `json:"int"`
	WIS   int `json:"wis"`
	FAITH int `json:"faith"`
}

// User is the character profile returned to the client (never includes the hash).
type User struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Title         string     `json:"title"`
	Level         int        `json:"level"`
	EXP           int        `json:"exp"`
	NextEXP       int        `json:"next_exp"`
	Coins         int        `json:"coins"`
	Streak        int        `json:"streak"`
	StreakFreezes int        `json:"streak_freezes"`
	Attr          Attributes `json:"attributes"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Category maps a habit type to the attribute it grows.
type Category struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Icon      string `json:"icon"`
	Attribute string `json:"attribute"`
}

// Quest is a habit definition plus this period's completion state.
type Quest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Category   string `json:"category"`
	Icon       string `json:"icon"`
	Attribute  string `json:"attribute"`
	Frequency  string `json:"frequency"`
	Difficulty string `json:"difficulty"`
	EXPReward  int    `json:"exp_reward"`
	CoinReward int    `json:"coin_reward"`
	Reminder   string `json:"reminder,omitempty"`
	// Scheduling: weekday (0=Sunday..6=Saturday) for weekly quests;
	// day_of_month (1..31) for monthly quests. Null otherwise.
	Weekday    *int   `json:"weekday,omitempty"`
	DayOfMonth *int   `json:"day_of_month,omitempty"`
	Schedule   string `json:"schedule"` // human label e.g. "Setiap Minggu", "Tanggal 15"
	Streak     int    `json:"streak"`
	Done       bool   `json:"done"`      // completed in the current period
	DueToday   bool   `json:"due_today"` // scheduled for today
}

// Comment on a feed post.
type Comment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// Post in the activity feed.
type Post struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Author      string    `json:"author"`
	AuthorLevel int       `json:"author_level"`
	Body        string    `json:"body"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	Badge       string    `json:"badge,omitempty"`
	Likes       int       `json:"likes"`
	LikedByMe   bool      `json:"liked_by_me"`
	Comments    []Comment `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
}

// Friend is a lightweight view of another user.
type Friend struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	Streak    int    `json:"streak"`
	WeeklyEXP int    `json:"weekly_exp"`
	IsMe      bool   `json:"is_me"`
}

// UserSearchResult is a candidate to add as a friend.
type UserSearchResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	Title     string `json:"title"`
	IsFriend  bool   `json:"is_friend"`
}

// Verse is a single Bible verse (book/chapter live on the response wrapper).
type Verse struct {
	Verse   int    `json:"verse"`
	TextID  string `json:"text_id"`
	TextEN  string `json:"text_en,omitempty"`
	Meaning string `json:"meaning,omitempty"`
}

// BibleBook is a canonical book plus the chapter numbers that have verses.
type BibleBook struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Ordinal   int     `json:"ordinal"`
	Testament string  `json:"testament"`
	Chapters  []int32 `json:"chapters"`
}

// Devotional is a daily reflection.
type Devotional struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Passage     string `json:"passage"`
	VerseText   string `json:"verse_text"`
	Reflection  string `json:"reflection"`
	Prayer      string `json:"prayer"`
	FaithReward int    `json:"faith_reward"`
	Completed   bool   `json:"completed"`
}

// Achievement is a badge with the user's live unlock state and progress.
type Achievement struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Hint     string `json:"hint"`
	Unlocked bool   `json:"unlocked"`
	Progress int    `json:"progress"`
	Target   int    `json:"target"`
}

// JournalEntry is one day's journal.
type JournalEntry struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"` // YYYY-MM-DD
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Mood      string    `json:"mood,omitempty"`
	Prompt    string    `json:"prompt,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AnalyticsDay is one day's activity for the progress charts.
type AnalyticsDay struct {
	Date        string `json:"date"` // YYYY-MM-DD
	Label       string `json:"label"`
	Completions int    `json:"completions"`
	EXP         int    `json:"exp"`
}

// CategoryCount is completions grouped by category.
type CategoryCount struct {
	Category  string `json:"category"`
	Attribute string `json:"attribute"`
	Count     int    `json:"count"`
}

// Analytics is the progress/history payload.
type Analytics struct {
	Daily            []AnalyticsDay  `json:"daily"`
	ByCategory       []CategoryCount `json:"by_category"`
	Attributes       Attributes      `json:"attributes"`
	TotalCompletions int             `json:"total_completions"`
	ActiveDays       int             `json:"active_days"`
	CurrentStreak    int             `json:"current_streak"`
	ThisWeekEXP      int             `json:"this_week_exp"`
}
