package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

var db *sql.DB

func main() {
	
	connStr := "host=localhost port=5432 user=postgres password=*** dbname=testinsta sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Database not reachable:", err)
	}
	fmt.Println("Connected to PostgreSQL")

	fs := http.FileServer(http.Dir("./fr"))
	http.Handle("/", fs)

	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    var u User
    err := json.NewDecoder(r.Body).Decode(&u)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    var dbPassword string
    err = db.QueryRow("SELECT password FROM users WHERE user_id=$1", u.UserID).Scan(&dbPassword)
    if err != nil {
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "error",
            "message": "User not found",
        })
        return
    }

    if u.Password == dbPassword {
        json.NewEncoder(w).Encode(map[string]string{"status": "success"})
    } else {
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "error",
            "message": "Incorrect password",
        })
    }
}


func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO users (user_id, password) VALUES ($1, $2)", u.UserID, u.Password)
	if err != nil {
		http.Error(w, "DB insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
