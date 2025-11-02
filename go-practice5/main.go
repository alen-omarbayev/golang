package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)


type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Genre  string `json:"genre"`
	Price  int    `json:"price"`
}

var db *pgxpool.Pool

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/library?sslmode=disable"
	}

	var err error
	ctx := context.Background()
	db, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/books", getBooksHandler)

	addr := ":8080"
	log.Printf("listening on %s ...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getBooksHandler(w http.ResponseWriter, r *http.Request) {
	
	q := r.URL.Query()

	genre := strings.TrimSpace(q.Get("genre"))
	sortParam := strings.TrimSpace(q.Get("sort")) 
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")


	limit := 10
	offset := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v >= 0 {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	
	const maxLimit = 100
	if limit > maxLimit {
		limit = maxLimit
	}

	
	var sb strings.Builder
	sb.WriteString("SELECT id, title, author, genre, price FROM books")
	args := make([]any, 0, 4)
	argPos := 1

	
	if genre != "" {
		sb.WriteString(fmt.Sprintf(" WHERE genre = $%d", argPos))
		args = append(args, genre)
		argPos++
	}

	
	if sortParam == "price_asc" {
		sb.WriteString(" ORDER BY price ASC")
	} else if sortParam == "price_desc" {
		sb.WriteString(" ORDER BY price DESC")
	}

	
	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, offset)

	query := sb.String()

	
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query, args...)
	elapsed := time.Since(start)
	
	w.Header().Set("X-Query-Time", fmt.Sprintf("%dms", elapsed.Milliseconds()))
	log.Printf("query took %s; sql=%q; args=%v", elapsed, query, args)

	if err != nil {
		http.Error(w, "database query error", http.StatusInternalServerError)
		log.Printf("db query error: %v", err)
		return
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.Price); err != nil {
			http.Error(w, "row scan error", http.StatusInternalServerError)
			log.Printf("row scan error: %v", err)
			return
		}
		books = append(books, b)
	}
	if rows.Err() != nil {
		http.Error(w, "rows error", http.StatusInternalServerError)
		log.Printf("rows error: %v", rows.Err())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
