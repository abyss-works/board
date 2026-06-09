package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//go:embed frontend/dist/*
var distFS embed.FS

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
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
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
	// 기존 테이블에 컬럼이 없으면 추가 (마이그레이션)
	db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP`)
	db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
	db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS password_hash TEXT`)
	log.Println("Database initialized successfully")
}

func main() {
	initDB()
	defer db.Close()

	http.Handle("/", spaHandler())
	http.HandleFunc("/api/posts", handlePosts)
	http.HandleFunc("/api/posts/", handlePostByID)
	http.HandleFunc("/api/comments", handleComments)

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func spaHandler() http.Handler {
	fsys, err := fs.Sub(distFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		f, err := fsys.Open(path)
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, content, author, created_at FROM posts WHERE deleted_at IS NULL ORDER BY created_at DESC")
		if err != nil {
			log.Printf("handlePosts GET: %v", err)
			http.Error(w, "Internal server error", 500)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)

	case "POST":
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		author := "나그네"
		var passwordHash *string
		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("handlePosts POST bcrypt: %v", err)
				http.Error(w, "Internal server error", 500)
				return
			}
			s := string(hash)
			passwordHash = &s
		}
		var p Post
		err := db.QueryRow(
			"INSERT INTO posts (title, content, author, password_hash) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
			req.Title, req.Content, author, passwordHash,
		).Scan(&p.ID, &p.CreatedAt)
		if err != nil {
			log.Printf("handlePosts POST: %v", err)
			http.Error(w, "Internal server error", 500)
			return
		}
		p.Title = req.Title
		p.Content = req.Content
		p.Author = author
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handlePostByID(w http.ResponseWriter, r *http.Request) {
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

	switch r.Method {
	case "GET":
		var p Post
		err := db.QueryRow("SELECT id, title, content, author, created_at FROM posts WHERE id=$1 AND deleted_at IS NULL", postID).
			Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", 404)
			return
		}
		if err != nil {
			log.Printf("handlePostByID GET: %v", err)
			http.Error(w, "Internal server error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)

	case "PUT":
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Title == "" || req.Content == "" {
			http.Error(w, "Title and content required", 400)
			return
		}
		// 비밀번호 검증
		if req.Password == "" {
			http.Error(w, "Password required", 401)
			return
		}
		var hash sql.NullString
		db.QueryRow("SELECT password_hash FROM posts WHERE id=$1 AND deleted_at IS NULL", postID).Scan(&hash)
		if !hash.Valid || bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.Password)) != nil {
			http.Error(w, "Invalid password", 401)
			return
		}
		result, err := db.Exec(
			"UPDATE posts SET title=$1, content=$2, updated_at=NOW() WHERE id=$3 AND deleted_at IS NULL",
			req.Title, req.Content, postID,
		)
		if err != nil {
			log.Printf("handlePostByID PUT: %v", err)
			http.Error(w, "Internal server error", 500)
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			http.Error(w, "Post not found", 404)
			return
		}
		var p Post
		db.QueryRow("SELECT id, title, content, author, created_at FROM posts WHERE id=$1", postID).
			Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)

	case "DELETE":
		var req struct {
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Password == "" {
			http.Error(w, "Password required", 401)
			return
		}
		var hash sql.NullString
		db.QueryRow("SELECT password_hash FROM posts WHERE id=$1 AND deleted_at IS NULL", postID).Scan(&hash)
		if !hash.Valid || bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.Password)) != nil {
			http.Error(w, "Invalid password", 401)
			return
		}
		result, err := db.Exec(
			"UPDATE posts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL",
			postID,
		)
		if err != nil {
			log.Printf("handlePostByID DELETE: %v", err)
			http.Error(w, "Internal server error", 500)
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			http.Error(w, "Post not found", 404)
			return
		}
		w.WriteHeader(204)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handlePostComments(w http.ResponseWriter, r *http.Request, postID int) {
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, post_id, content, author, created_at FROM comments WHERE post_id=$1 ORDER BY created_at ASC", postID)
		if err != nil {
			log.Printf("handlePostComments GET: %v", err)
			http.Error(w, "Internal server error", 500)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)

	case "POST":
		var c Comment
		json.NewDecoder(r.Body).Decode(&c)
		c.PostID = postID
		c.Author = "나그네"
		err := db.QueryRow(
			"INSERT INTO comments (post_id, content, author) VALUES ($1, $2, $3) RETURNING id, created_at",
			c.PostID, c.Content, c.Author,
		).Scan(&c.ID, &c.CreatedAt)
		if err != nil {
			log.Printf("handlePostComments POST: %v", err)
			http.Error(w, "Internal server error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(c)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleComments(w http.ResponseWriter, r *http.Request) {
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
