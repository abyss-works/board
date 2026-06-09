package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func initDB() {
	host := getEnv("DB_HOST", "postgres")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "community")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for PostgreSQL... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			author TEXT NOT NULL DEFAULT 'Anonymous',
			created_at TIMESTAMP DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			author TEXT NOT NULL DEFAULT 'Anonymous',
			created_at TIMESTAMP DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Database initialized successfully")
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/posts", handlePosts)
	http.HandleFunc("/api/posts/", handlePostByID)
	http.HandleFunc("/api/comments", handleComments)

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	tmpl.Execute(w, nil)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, content, author, created_at FROM posts ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var p Post
			rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
			posts = append(posts, p)
		}
		if posts == nil {
			posts = []Post{}
		}
		json.NewEncoder(w).Encode(posts)

	case "POST":
		var p Post
		json.NewDecoder(r.Body).Decode(&p)
		if p.Author == "" {
			p.Author = "Anonymous"
		}
		err := db.QueryRow(
			"INSERT INTO posts (title, content, author) VALUES ($1, $2, $3) RETURNING id, created_at",
			p.Title, p.Content, p.Author,
		).Scan(&p.ID, &p.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handlePostByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from /api/posts/123 or /api/posts/123/comments
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Post ID required", 400)
		return
	}

	postID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid post ID", 400)
		return
	}

	if len(parts) >= 2 && parts[1] == "comments" {
		handlePostComments(w, r, postID)
		return
	}

	// GET single post (not really used by frontend but nice to have)
	if r.Method == "GET" {
		var p Post
		err := db.QueryRow("SELECT id, title, content, author, created_at FROM posts WHERE id=$1", postID).
			Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		if err != nil {
			http.Error(w, "Post not found", 404)
			return
		}
		json.NewEncoder(w).Encode(p)
		return
	}

	http.Error(w, "Method not allowed", 405)
}

func handlePostComments(w http.ResponseWriter, r *http.Request, postID int) {
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, post_id, content, author, created_at FROM comments WHERE post_id=$1 ORDER BY created_at ASC", postID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var comments []Comment
		for rows.Next() {
			var c Comment
			rows.Scan(&c.ID, &c.PostID, &c.Content, &c.Author, &c.CreatedAt)
			comments = append(comments, c)
		}
		if comments == nil {
			comments = []Comment{}
		}
		json.NewEncoder(w).Encode(comments)

	case "POST":
		var c Comment
		json.NewDecoder(r.Body).Decode(&c)
		c.PostID = postID
		if c.Author == "" {
			c.Author = "Anonymous"
		}
		err := db.QueryRow(
			"INSERT INTO comments (post_id, content, author) VALUES ($1, $2, $3) RETURNING id, created_at",
			c.PostID, c.Content, c.Author,
		).Scan(&c.ID, &c.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(c)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	// Alternate endpoint: /api/comments?post_id=123
	if r.Method == "GET" {
		postIDStr := r.URL.Query().Get("post_id")
		if postIDStr == "" {
			http.Error(w, "post_id query parameter required", 400)
			return
		}
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid post_id", 400)
			return
		}
		handlePostComments(w, r, postID)
		return
	}
	http.Error(w, "Method not allowed", 405)
}

const indexHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>커뮤니티</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@300;400;500&family=Noto+Serif+KR:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
/* ===== Theme Variables ===== */
:root {
  --bg: #f5f5f0;
  --surface: #efefeb;
  --text: #1a1a1a;
  --text-secondary: #5c5c5c;
  --text-muted: #8c8c8c;
  --border: #e5e5e0;
  --border-light: #eaeae5;
  --accent: #6b6b6b;
  --accent-hover: #555555;
  --accent-light: #f0f0eb;
  --accent-soft: rgba(59, 59, 59, 0.06);
  --success: #5b8c5a;
  --danger: #c25450;
  --btn-bg: #3b3b3b;
  --btn-hover: #2a2a2a;
  --btn-text: #ffffff;
  --shadow: none;
  --shadow-md: 0 1px 3px rgba(0,0,0,0.03);
  --shadow-lg: 0 4px 16px rgba(0,0,0,0.05);
  --radius: 8px;
  --radius-sm: 6px;
  --transition: 0.15s ease;
  --input-bg: #efefeb;
  --focus-ring: 0 0 0 2px rgba(59, 59, 59, 0.12);
}
[data-theme="dark"] {
  --bg: #2d2d2d;
  --surface: #363636;
  --text: #e0e0e0;
  --text-secondary: #a0a0a0;
  --text-muted: #707070;
  --border: #444444;
  --border-light: #3a3a3a;
  --accent: #8c8c8c;
  --accent-hover: #a0a0a0;
  --accent-light: #333333;
  --accent-soft: rgba(255,255,255,0.05);
  --success: #5b8c5a;
  --danger: #c25450;
  --btn-bg: #4a4a4a;
  --btn-hover: #3a3a3a;
  --btn-text: #e0e0e0;
  --shadow: none;
  --shadow-md: 0 1px 3px rgba(0,0,0,0.15);
  --shadow-lg: 0 4px 16px rgba(0,0,0,0.2);
  --radius: 8px;
  --radius-sm: 6px;
  --transition: 0.15s ease;
  --input-bg: #363636;
  --focus-ring: 0 0 0 2px rgba(255,255,255,0.1);
}

/* ===== Reset & Base ===== */
*, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
html { scroll-behavior: smooth; -webkit-font-smoothing: antialiased; }
body {
  font-family: 'Noto Sans KR', -apple-system, BlinkMacSystemFont, sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.7;
  font-size: 16px;
  transition: background 0.3s ease, color 0.3s ease;
  min-height: 100vh;
}

/* ===== Layout ===== */
.container {
  max-width: 680px;
  margin: 0 auto;
  padding: 48px 24px 80px;
}

/* ===== Header ===== */
.header {
  text-align: center;
  padding: 40px 0 48px;
  position: relative;
}
.site-title {
  font-family: 'Noto Serif KR', 'Georgia', serif;
  font-size: 1.8rem;
  font-weight: 600;
  color: var(--text);
  letter-spacing: -0.01em;
  margin-bottom: 8px;
}
.title-ornament {
  color: var(--text-muted);
  font-size: 0.5rem;
  letter-spacing: 12px;
  margin-bottom: 8px;
  user-select: none;
  opacity: 0.5;
}
.site-desc {
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-weight: 300;
  letter-spacing: 0.02em;
}
.theme-btn {
  position: absolute;
  top: 40px;
  right: 0;
  background: none;
  border: 1px solid var(--border);
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--text-muted);
  padding: 5px 12px;
  border-radius: 6px;
  font-family: inherit;
  transition: all var(--transition);
  letter-spacing: 0.03em;
}
.theme-btn:hover {
  color: var(--text);
  border-color: var(--text-muted);
}

/* ===== Section Heading ===== */
.section-heading {
  font-family: 'Noto Serif KR', serif;
  font-size: 1rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin: 0 0 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border);
  letter-spacing: -0.01em;
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.section-heading .count-badge {
  font-family: 'Noto Sans KR', sans-serif;
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--text-muted);
}

/* ===== New Post Section ===== */
.new-post-section {
  margin-bottom: 44px;
}
.new-post-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: border-color var(--transition);
}
.new-post-card:hover {
  border-color: var(--text-muted);
}
.form-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  cursor: pointer;
  user-select: none;
  font-family: 'Noto Serif KR', serif;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--text);
  letter-spacing: -0.01em;
}
.form-toggle .toggle-arrow {
  font-size: 0.7rem;
  color: var(--text-muted);
  transition: transform 0.25s ease;
}
.new-post-card.collapsed .toggle-arrow { transform: rotate(-90deg); }
.new-post-card.collapsed .form-body { display: none; }
.form-body { padding: 0 20px 24px; }
.form-group { margin-bottom: 18px; }
.form-group label {
  display: block;
  font-family: 'Noto Sans KR', sans-serif;
  font-size: 0.8rem;
  font-weight: 500;
  margin-bottom: 6px;
  color: var(--text-secondary);
  letter-spacing: 0.02em;
}
.form-group input,
.form-group textarea {
  width: 100%;
  padding: 10px 14px;
  border: 0;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-family: 'Noto Sans KR', sans-serif;
  background: var(--input-bg);
  color: var(--text);
  transition: all var(--transition);
  outline: none;
  line-height: 1.6;
}
.form-group input:focus,
.form-group textarea:focus {
  box-shadow: var(--focus-ring);
}
.form-group textarea { resize: vertical; min-height: 110px; }
.form-group input::placeholder,
.form-group textarea::placeholder { color: var(--text-muted); font-weight: 300; }

/* ===== Buttons ===== */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 22px;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  font-family: 'Noto Sans KR', sans-serif;
  cursor: pointer;
  border: none;
  transition: background var(--transition);
  white-space: nowrap;
  letter-spacing: 0.01em;
}
.btn-primary {
  background: var(--btn-bg);
  color: var(--btn-text);
}
.btn-primary:hover {
  background: var(--btn-hover);
}
.btn-primary:active { transform: scale(0.98); }
.btn-sm { padding: 6px 14px; font-size: 0.8rem; border-radius: 6px; }

/* ===== Post Cards ===== */
.post-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 22px 24px;
  margin-bottom: 10px;
  transition: border-color var(--transition);
  animation: fadeInUp 0.4s ease both;
  position: relative;
}
.post-card:hover {
  border-color: var(--text-muted);
}
.post-title {
  font-family: 'Noto Serif KR', serif;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 8px;
  line-height: 1.45;
  letter-spacing: -0.01em;
}
.post-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.post-meta .meta-sep { color: var(--border); font-weight: 300; }
.post-meta .post-id {
  font-family: 'Noto Serif KR', serif;
  font-size: 0.7rem;
  color: var(--text-muted);
  opacity: 0.5;
}
.post-preview {
  font-size: 0.92rem;
  line-height: 1.7;
  color: var(--text-secondary);
  margin-bottom: 14px;
  white-space: pre-wrap;
  word-break: break-word;
}
.comment-toggle {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--text-muted);
  font-family: 'Noto Sans KR', sans-serif;
  padding: 4px 0;
  transition: color var(--transition);
}
.comment-toggle:hover { color: var(--text); }
.comment-toggle .toggle-count {
  font-size: 0.7rem;
  opacity: 0.7;
}

/* ===== Comments ===== */
.comments-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border);
  animation: fadeIn 0.3s ease;
}
.comment-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border-light);
  animation: fadeIn 0.25s ease both;
}
.comment-item:last-child { border-bottom: none; }
.comment-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 0.78rem;
  color: var(--text-muted);
}
.comment-avatar {
  width: 22px; height: 22px;
  border-radius: 50%;
  background: var(--border-light);
  color: var(--text-muted);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.6rem; font-weight: 600;
  font-family: 'Noto Serif KR', serif;
  flex-shrink: 0;
}
.comment-author { font-weight: 500; color: var(--text-secondary); }
.comment-date { font-size: 0.72rem; }
.comment-body {
  font-size: 0.9rem;
  line-height: 1.65;
  color: var(--text);
  padding-left: 30px;
  word-break: break-word;
}
.empty-comments {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
  padding: 24px 0;
  font-weight: 300;
}
.comment-form {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid var(--border);
}
.comment-form input {
  width: 100%;
  padding: 9px 14px;
  border: 0;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-family: 'Noto Sans KR', sans-serif;
  background: var(--input-bg);
  color: var(--text);
  margin-bottom: 8px;
  outline: none;
  transition: box-shadow var(--transition);
}
.comment-form input:focus { box-shadow: var(--focus-ring); }
.comment-form .comment-form-row {
  display: flex;
  gap: 8px;
}
.comment-form .comment-form-row input {
  flex: 1;
  margin-bottom: 0;
}

/* ===== States ===== */
.loading {
  text-align: center;
  color: var(--text-muted);
  padding: 64px 20px;
  font-weight: 300;
  font-size: 0.9rem;
}
.loading-spinner {
  width: 28px; height: 28px;
  border: 2px solid var(--border);
  border-top-color: var(--text-muted);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}
.error {
  text-align: center;
  color: var(--danger);
  padding: 40px 20px;
  background: rgba(194, 84, 80, 0.04);
  border-radius: var(--radius);
  border: 1px solid rgba(194, 84, 80, 0.12);
  font-weight: 300;
}
.error small { font-size: 0.8rem; opacity: 0.7; }
.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 72px 20px;
  font-weight: 300;
}
.empty-icon { font-size: 2.5rem; margin-bottom: 16px; opacity: 0.35; }
.empty p { margin-top: 6px; font-size: 0.95rem; }

/* ===== Toast ===== */
.toast {
  position: fixed;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%) translateY(120px);
  background: var(--text);
  color: var(--bg);
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 400;
  box-shadow: var(--shadow-lg);
  z-index: 1000;
  transition: transform 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  pointer-events: none;
  font-family: 'Noto Sans KR', sans-serif;
  letter-spacing: 0.01em;
}
.toast.show { transform: translateX(-50%) translateY(0); }
.toast.success { background: var(--success); color: #fff; }
.toast.error { background: var(--danger); color: #fff; }

/* ===== Animations ===== */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== Responsive ===== */
@media (max-width: 640px) {
  .container { padding: 28px 16px 60px; }
  .header { padding: 24px 0 32px; }
  .site-title { font-size: 1.5rem; }
  .theme-btn { top: 24px; font-size: 0.75rem; padding: 4px 10px; }
  .post-card { padding: 18px 18px; }
  .post-title { font-size: 1rem; }
  .post-preview { font-size: 0.88rem; }
  .form-body { padding: 0 16px 20px; }
  .form-toggle { padding: 14px 16px; font-size: 0.9rem; }
  .btn { padding: 8px 18px; font-size: 0.82rem; }
  .section-heading { font-size: 0.95rem; }
}

@media (max-width: 380px) {
  .site-title { font-size: 1.3rem; }
  .comment-form .comment-form-row { flex-direction: column; }
  .btn { width: 100%; }
  .post-card { padding: 14px 14px; }
}
</style>
</head>
<body>
<div class="container">
  <!-- Header -->
  <header class="header">
    <button class="theme-btn" onclick="toggleTheme()" id="themeToggle" aria-label="테마 전환">🌙</button>
    <h1 class="site-title">커뮤니티</h1>
    <div class="title-ornament">&mdash; &sect; &mdash;</div>
    <p class="site-desc">생각을 나누는 공간</p>
  </header>

  <!-- New Post -->
  <section class="new-post-section">
    <div class="new-post-card" id="newPostCard">
      <div class="form-toggle" onclick="toggleNewPost()">
        <span>새 글 작성</span>
        <span class="toggle-arrow">&#9660;</span>
      </div>
      <div class="form-body">
        <form id="postForm">
          <div class="form-group">
            <label>제목</label>
            <input type="text" id="postTitle" required placeholder="무슨 이야기를 들려주시겠어요?">
          </div>
          <div class="form-group">
            <label>내용</label>
            <textarea id="postContent" required placeholder="자유롭게 생각을 풀어내 보세요."></textarea>
          </div>
          <div class="form-group">
            <label>작성자</label>
            <input type="text" id="postAuthor" placeholder="익명">
          </div>
          <button type="submit" class="btn btn-primary">등록하기</button>
        </form>
      </div>
    </div>
  </section>

  <!-- Posts -->
  <section>
    <h2 class="section-heading">
      글 목록
      <span class="count-badge" id="postCount" style="display:none"></span>
    </h2>
    <div id="postsContainer">
      <div class="loading">
        <div class="loading-spinner"></div>
        <div>글을 불러오는 중&hellip;</div>
      </div>
    </div>
  </section>
</div>

<!-- Toast -->
<div class="toast" id="toast"></div>

<script>
const API = '/api';

// ===== Theme Toggle =====
function toggleTheme() {
  const html = document.documentElement;
  const btn = document.getElementById('themeToggle');
  const isDark = html.getAttribute('data-theme') === 'dark';
  html.setAttribute('data-theme', isDark ? 'light' : 'dark');
  btn.textContent = isDark ? '\uD83C\uDF19' : '\u2600\uFE0F';
  localStorage.setItem('theme', isDark ? 'light' : 'dark');
}

(function initTheme() {
  const saved = localStorage.getItem('theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  document.getElementById('themeToggle').textContent = saved === 'dark' ? '\u2600\uFE0F' : '\uD83C\uDF19';
})();

// ===== Collapse New Post =====
function toggleNewPost() {
  document.getElementById('newPostCard').classList.toggle('collapsed');
}

// ===== Toast =====
var toastTimer;
function showToast(msg, type) {
  type = type || '';
  var toast = document.getElementById('toast');
  toast.textContent = msg;
  toast.className = 'toast ' + type + ' show';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(function() { toast.classList.remove('show'); }, 2600);
}

// ===== Load Posts =====
async function loadPosts() {
  var container = document.getElementById('postsContainer');
  var countBadge = document.getElementById('postCount');
  try {
    var res = await fetch(API + '/posts');
    var posts = await res.json();
    if (posts.length === 0) {
      container.innerHTML = '<div class="empty"><div class="empty-icon">&#128238;</div><p>아직 올라온 글이 없습니다.</p><p style="margin-top:4px;font-size:0.85rem">가장 먼저 이야기를 시작해 보세요.</p></div>';
      countBadge.style.display = 'none';
      return;
    }
    countBadge.textContent = '(' + posts.length + ')';
    countBadge.style.display = 'inline';
    var html = '';
    for (var i = 0; i < posts.length; i++) {
      var post = posts[i];
      var date = new Date(post.created_at).toLocaleString('ko-KR', {
        year: 'numeric', month: 'long', day: 'numeric',
        hour: '2-digit', minute: '2-digit'
      });
      var preview = post.content.length > 180
        ? escapeHtml(post.content).substring(0, 180) + '&hellip;'
        : escapeHtml(post.content);
      html += '<div class="post-card" style="animation-delay:' + (i * 0.04) + 's">' +
        '<div class="post-title">' + escapeHtml(post.title) + '</div>' +
        '<div class="post-meta">' +
          '<span>' + escapeHtml(post.author) + '</span>' +
          '<span class="meta-sep">&middot;</span>' +
          '<span>' + date + '</span>' +
          '<span class="meta-sep">&middot;</span>' +
          '<span class="post-id">#' + post.id + '</span>' +
        '</div>' +
        '<div class="post-preview">' + preview + '</div>' +
        '<button class="comment-toggle" onclick="toggleComments(' + post.id + ', this)">' +
          '&#9998; 댓글' +
        '</button>' +
        '<div id="comments-' + post.id + '" class="comments-section" style="display:none"></div>' +
      '</div>';
    }
    container.innerHTML = html;
  } catch(e) {
    container.innerHTML = '<div class="error">&#9888; 글을 불러오지 못했습니다.<br><small>잠시 후 다시 시도해 주세요.</small></div>';
  }
}

// ===== Toggle Comments =====
async function toggleComments(postId, btn) {
  var container = document.getElementById('comments-' + postId);
  if (container.style.display === 'none' || !container.style.display) {
    container.style.display = 'block';
    btn.innerHTML = '&#9998; 댓글 접기';
    btn.style.color = 'var(--accent)';
    if (!container.dataset.loaded) {
      container.dataset.loaded = '1';
      await loadComments(postId, container);
    }
  } else {
    container.style.display = 'none';
    btn.innerHTML = '&#9998; 댓글';
    btn.style.color = '';
  }
}

// ===== Load Comments =====
async function loadComments(postId, container) {
  try {
    var res = await fetch(API + '/posts/' + postId + '/comments');
    var comments = await res.json();
    var html = '';
    if (comments.length === 0) {
      html += '<div class="empty-comments">아직 댓글이 없습니다. 처음으로 의견을 남겨보세요.</div>';
    } else {
      for (var i = 0; i < comments.length; i++) {
        var c = comments[i];
        var date = new Date(c.created_at).toLocaleString('ko-KR', {
          month: 'long', day: 'numeric',
          hour: '2-digit', minute: '2-digit'
        });
        var initial = escapeHtml(c.author).charAt(0).toUpperCase();
        html += '<div class="comment-item" style="animation-delay:' + (i * 0.03) + 's">' +
          '<div class="comment-meta">' +
            '<span class="comment-avatar">' + initial + '</span>' +
            '<span class="comment-author">' + escapeHtml(c.author) + '</span>' +
            '<span class="meta-sep">&middot;</span>' +
            '<span class="comment-date">' + date + '</span>' +
          '</div>' +
          '<div class="comment-body">' + escapeHtml(c.content) + '</div>' +
        '</div>';
      }
    }
    html += '<div class="comment-form">' +
      '<form onsubmit="addComment(event, ' + postId + ')">' +
        '<input type="text" id="commentAuthor-' + postId + '" placeholder="이름 (선택)">' +
        '<div class="comment-form-row">' +
          '<input type="text" id="commentContent-' + postId + '" placeholder="댓글을 입력하세요&hellip;" required>' +
          '<button type="submit" class="btn btn-primary btn-sm">등록</button>' +
        '</div>' +
      '</form></div>';
    container.innerHTML = html;
  } catch(e) {
    container.innerHTML = '<div class="error" style="padding:20px;font-size:0.85rem">&#9888; 댓글을 불러오지 못했습니다.</div>';
  }
}

// ===== Add Comment =====
async function addComment(e, postId) {
  e.preventDefault();
  var content = document.getElementById('commentContent-' + postId).value;
  var author = document.getElementById('commentAuthor-' + postId).value || '\uC775\uBA85';
  try {
    var res = await fetch(API + '/posts/' + postId + '/comments', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({content: content, author: author})
    });
    if (!res.ok) throw new Error('Failed');
    var container = document.getElementById('comments-' + postId);
    await loadComments(postId, container);
    showToast('\u2714 \uB313\uAE00\uC774 \uB4F1\uB85D\uB418\uC5C8\uC2B5\uB2C8\uB2E4', 'success');
  } catch(e) {
    showToast('\u2716 \uB313\uAE00 \uB4F1\uB85D\uC5D0 \uC2E4\uD328\uD588\uC2B5\uB2C8\uB2E4', 'error');
  }
}

// ===== Submit Post =====
document.getElementById('postForm').addEventListener('submit', async function(e) {
  e.preventDefault();
  var title = document.getElementById('postTitle').value.trim();
  var content = document.getElementById('postContent').value.trim();
  var author = document.getElementById('postAuthor').value.trim() || '\uC775\uBA85';

  if (!title || !content) {
    showToast('\u26A0 \uC81C\uBAA9\uACFC \uB0B4\uC6A9\uC744 \uBAA8\uB450 \uC785\uB825\uD574\uC8FC\uC138\uC694', 'error');
    return;
  }

  var submitBtn = this.querySelector('button[type="submit"]');
  var originalText = submitBtn.textContent;
  submitBtn.textContent = '\uB4F1\uB85D \uC911\u2026';
  submitBtn.disabled = true;

  try {
    var res = await fetch(API + '/posts', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({title: title, content: content, author: author})
    });
    if (!res.ok) throw new Error('Failed');
    document.getElementById('postForm').reset();
    await loadPosts();
    showToast('\u2714 \uAE00\uC774 \uB4F1\uB85D\uB418\uC5C8\uC2B5\uB2C8\uB2E4', 'success');
    if (window.innerWidth < 640) {
      document.getElementById('newPostCard').classList.add('collapsed');
    }
  } catch(e) {
    showToast('\u2716 \uAE00 \uB4F1\uB85D\uC5D0 \uC2E4\uD328\uD588\uC2B5\uB2C8\uB2E4', 'error');
  } finally {
    submitBtn.textContent = originalText;
    submitBtn.disabled = false;
  }
});

// ===== Escape HTML =====
function escapeHtml(text) {
  if (!text) return '';
  var div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// ===== Initial Load =====
loadPosts();
</script>
</body>
</html>`
