package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config represents the application configuration
type Config struct {
	Backend struct {
		Port int `json:"port"`
	} `json:"backend"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"database"`
}

// Database connection
var db *sql.DB
var config Config

// Entry represents a password entry
type Entry struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Remark   string `json:"remark"`
}

// Log represents an operation log
type Log struct {
	ID            int       `json:"id"`
	EntryID       int       `json:"entry_id"`
	OperationType string    `json:"operation_type"`
	OperationTime time.Time `json:"operation_time"`
	UserID        int       `json:"user_id"`
}

func loadConfig() error {
	configFile, err := os.ReadFile("../config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(configFile, &config)
}

func initDB() {
	var err error
	dsn := config.Database.User + ":" + config.Database.Password + "@tcp(" + config.Database.Host + ":" + strconv.Itoa(config.Database.Port) + ")/" + config.Database.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database")
}

func main() {
	initDB()
	defer db.Close()

	// API routes
	http.HandleFunc("/api/password_entries", entriesHandler)
	http.HandleFunc("/api/password_entries/", entryHandler)
	http.HandleFunc("/api/operation_logs", logsHandler)

	// Serve static files from the correct path
	frontendPath := filepath.Join(".", "..", "frontend")
	absPath, _ := filepath.Abs(frontendPath)
	log.Println("Serving static files from:", absPath)

	// Check if frontend directory exists and is readable
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		log.Fatal("Frontend directory not found:", frontendPath)
	}
	if _, err := os.Stat(filepath.Join(frontendPath, "index.html")); os.IsNotExist(err) {
		log.Fatal("index.html not found in frontend directory")
	}

	// Create file server with logging
	fs := http.FileServer(http.Dir(frontendPath))

	// Handle all routes by first trying static files, then falling back to index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log the requested path
		log.Printf("Request path: %s", r.URL.Path)

		// Try to serve static file
		if _, err := os.Stat(filepath.Join(frontendPath, r.URL.Path)); os.IsNotExist(err) {
			// If file doesn't exist, serve index.html for SPA routing
			http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
			return
		}

		// Otherwise serve the static file
		http.StripPrefix("/", fs).ServeHTTP(w, r)
	})

	if err := loadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Starting server on :%d", config.Backend.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Backend.Port), nil))
}

// entriesHandler handles all password entry operations
func entriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getEntries(w, r)
	case "POST":
		createEntry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// entryHandler handles single entry operations
func entryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getEntry(w, r)
	case "PUT":
		updateEntry(w, r)
	case "DELETE":
		deleteEntry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// logsHandler handles operation log requests
func logsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	getLogs(w, r)
}

// getEntries returns all password entries
func getEntries(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, username, password, remark FROM password_entries")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Title, &e.Username, &e.Password, &e.Remark); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		entries = append(entries, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// getEntry returns a single password entry
func getEntry(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/password_entries/"):]
	var entry Entry

	err := db.QueryRow("SELECT id, title, username, password, remark FROM password_entries WHERE id = ?", id).
		Scan(&entry.ID, &entry.Title, &entry.Username, &entry.Password, &entry.Remark)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Entry not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// createEntry adds a new password entry
func createEntry(w http.ResponseWriter, r *http.Request) {
	var entry Entry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement password encryption
	result, err := db.Exec("INSERT INTO password_entries (title, username, password, remark) VALUES (?, ?, ?, ?)",
		entry.Title, entry.Username, entry.Password, entry.Remark)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	entry.ID = int(id)

	// Log the operation
	logOperation(entry.ID, "add", 1)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// updateEntry modifies an existing password entry
func updateEntry(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/password_entries/"):]
	var entry Entry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE password_entries SET title = ?, username = ?, password = ?, remark = ? WHERE id = ?",
		entry.Title, entry.Username, entry.Password, entry.Remark, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the operation
	logOperation(entry.ID, "update", 1)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// deleteEntry removes a password entry
func deleteEntry(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/password_entries/"):]

	// Log the operation before deletion
	logOperation(id, "delete", 1)

	_, err := db.Exec("DELETE FROM password_entries WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getLogs returns all operation logs
func getLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, entry_id, operation_type, operation_time, user_id FROM operation_logs")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.EntryID, &l.OperationType, &l.OperationTime, &l.UserID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// logOperation records an operation in the logs
func logOperation(entryID interface{}, operationType string, userID int) {
	_, err := db.Exec("INSERT INTO operation_logs (entry_id, operation_type, user_id) VALUES (?, ?, ?)",
		entryID, operationType, userID)
	if err != nil {
		log.Printf("Failed to log operation: %v", err)
	}
}
