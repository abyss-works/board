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
