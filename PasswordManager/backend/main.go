package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database connection
var db *sql.DB

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
	ID           int       `json:"id"`
	EntryID      int       `json:"entry_id"`
	OperationType string   `json:"operation_type"`
	OperationTime time.Time `json:"operation_time"`
	UserID       int       `json:"user_id"`
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/password_manager")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
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

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
