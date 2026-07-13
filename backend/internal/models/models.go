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
	Onboarded     bool       `json:"onboarded"`
	Locale        string     `json:"locale"`
	AvatarURL     string     `json:"avatar_url"`
	Timezone      string     `json:"timezone"`
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
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}

// Post in the activity feed.
type Post struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"author_avatar,omitempty"`
	AuthorLevel  int       `json:"author_level"`
	Body         string    `json:"body"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	Badge        string    `json:"badge,omitempty"`
	Likes        int       `json:"likes"`
	LikedByMe    bool      `json:"liked_by_me"`
	Comments     []Comment `json:"comments"`
	CreatedAt    time.Time `json:"created_at"`
}

// Friend is a lightweight view of another user.
type Friend struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	Level     int    `json:"level"`
	Streak    int    `json:"streak"`
	WeeklyEXP int    `json:"weekly_exp"`
	IsMe      bool   `json:"is_me"`
}

// UserSearchResult is a candidate to add as a friend.
// Status is one of: "none" | "outgoing" | "incoming" | "friend".
type UserSearchResult struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
	Level  int    `json:"level"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// FriendRequest is an incoming friend request (from another user).
type FriendRequest struct {
	ID     string `json:"id"` // requester's user id
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
	Level  int    `json:"level"`
	Title  string `json:"title"`
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

// BibleMark is a highlight or bookmark on a verse.
type BibleMark struct {
	BookID  string `json:"book_id"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Kind    string `json:"kind"` // "highlight" | "bookmark"
}

// Bookmark is a saved verse with its text, for the bookmarks list.
type Bookmark struct {
	BookID  string `json:"book_id"`
	Book    string `json:"book"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Text    string `json:"text"`
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

// Reading is one chapter to read within a plan day.
type Reading struct {
	BookID  string `json:"book_id"`
	Chapter int    `json:"chapter"`
	Label   string `json:"label"`
}

// ReadingPlanDay is one day of a reading plan.
type ReadingPlanDay struct {
	Day       int       `json:"day"`
	Label     string    `json:"label"`
	Readings  []Reading `json:"readings"`
	Completed bool      `json:"completed"`
}

// ReadingPlan is a Bible reading plan plus the user's progress.
type ReadingPlan struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"`
	TotalDays   int              `json:"total_days"`
	Enrolled    bool             `json:"enrolled"`
	Completed   int              `json:"completed"`
	FaithReward int              `json:"faith_reward"`
	Days        []ReadingPlanDay `json:"days,omitempty"`
}

// Prayer is a prayer-list entry.
type Prayer struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Answered   bool       `json:"answered"`
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// HeatmapDay is one calendar day's completion count.
type HeatmapDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
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
