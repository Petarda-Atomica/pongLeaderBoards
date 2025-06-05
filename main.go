package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Register SQLite driver
)

var remote_ip string
var score int
var qr_code []byte

var scoresDB *sql.DB

func main() {
	// Start broadcasting local IP over UDP
	go broadcastLocalIP(4210, time.Second*2)

	// Open scores database
	var err error
	scoresDB, err = sql.Open("sqlite", "scores.db")
	if err != nil {
		log.Fatal("Failed to open scores database:", err)
	}
	defer scoresDB.Close()

	// Make sure the table for the scores exists
	_, err = scoresDB.Exec(`
	CREATE TABLE IF NOT EXISTS Scores (
    	ID TEXT NOT NULL PRIMARY KEY,
    	user_name TEXT,
    	score INTEGER NOT NULL
	);`)
	if err != nil {
		log.Fatal("Failed to create scores table:", err)
	}

	// Start bore tunnel
	remote_ip, err = bore("4211")
	if err != nil {
		log.Fatal("Failed to start bore tunnel:", err)
	}
	log.Println("Forwarded to", remote_ip)

	// Create a blank qr
	blank_qr()

	// Handle verification
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I am leaderboard")
	})

	// Handle scoring
	http.HandleFunc("/score", func(w http.ResponseWriter, r *http.Request) {
		score += 1
		log.Println("Score:", score)
		fmt.Fprintf(w, "OK")
	})

	// Handle publishing scores
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		// Create a random unique ID for this score
		codeID, err := generateID(32)
		if err != nil {
			log.Println("Failed to generate an unique ID")
			log.Println(err)
			claim_score_qr(":(") // This will make sure we don't grab the wrong QR for the claim
			fmt.Fprint(w, "SERVER ERR: NO ID")
			return
		}

		// Generate a qrcode for the user
		claim_score_qr(codeID)

		// Create db request
		db_request := `
		INSERT INTO Scores (ID, score)
		VALUES ('{ID}', {SCORE});`
		db_request = strings.ReplaceAll(db_request, "{ID}", codeID)
		db_request = strings.ReplaceAll(db_request, "{SCORE}", fmt.Sprint(score))

		// Add the score in the database
		_, err = scoresDB.Exec(db_request)
		if err != nil {
			log.Println("Failed to register score")
			log.Println(err)
			fmt.Fprint(w, "SERVER ERR: DB ERR")
			return
		}

		// Reset score
		score = 0

		// Write back to the sender
		fmt.Fprintf(w, "OK")
	})

	// Handle claiming scores
	http.HandleFunc("/claim", func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		query := r.URL.Query()
		codeID := query.Get("code")

		// Get HTML document
		HTML_raw, err := os.ReadFile("html/claim.html")
		if err != nil {
			fmt.Fprint(w, "<html><body></h1>There was an unexpected Error!</h1></body></html>")
			return
		}

		// Insert variables
		HTML := string(HTML_raw)
		HTML = strings.ReplaceAll(HTML, "{REMOTE-IP}", remote_ip)
		HTML = strings.ReplaceAll(HTML, "{CODE}", codeID)

		// Send HTML
		fmt.Fprint(w, HTML)
	})

	// Handle rank webpage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HTML_raw, err := os.ReadFile("html/root.html")
		if err != nil {
			fmt.Fprint(w, "<html><body></h1>There was an unexpected Error!</h1></body></html>")
			return
		}
		HTML := string(HTML_raw)
		HTML = strings.ReplaceAll(HTML, "{REMOTE-IP}", remote_ip)

		fmt.Fprint(w, HTML)
	})

	// Handle score-display webpage
	http.HandleFunc("/score-display", func(w http.ResponseWriter, r *http.Request) {
		HTML_raw, err := os.ReadFile("html/score-display.html")
		if err != nil {
			fmt.Fprint(w, "<html><body></h1>There was an unexpected Error!</h1></body></html>")
			return
		}
		HTML := string(HTML_raw)
		HTML = strings.ReplaceAll(HTML, "{REMOTE-IP}", remote_ip)

		fmt.Fprint(w, HTML)
	})

	// Handle websocket communication
	http.HandleFunc("/ws", handleWebSocket)

	// Start webserver
	if err := http.ListenAndServe(":4211", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
