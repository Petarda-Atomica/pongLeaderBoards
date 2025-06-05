package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accept requests from any origin
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Wait for messages
	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		// Handle get_scores
		if string(message) == "get_scores" {
			ws_get_scores(conn)
		} else if string(message) == "get_current_score" {
			if score > 0 {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprint(score)))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString(qr_code)))
			}
		} else if strings.HasPrefix(string(message), ">") {
			log.Println(string(message))
			var elems []string
			err := json.Unmarshal([]byte(strings.TrimPrefix(string(message), ">")), &elems)
			if err != nil {
				log.Println("Failed to unmarshal json")
				log.Println(err)
				continue
			}
			name := elems[0]
			ID := elems[1]

			// Prepare update
			stmt, err := scoresDB.Prepare("UPDATE Scores SET user_name = ? WHERE ID = ?")
			if err != nil {
				log.Println("Failed to prepare SQL update")
				log.Println(err)
				continue
			}

			res, err := stmt.Exec(name, ID)
			if err != nil {
				log.Println("Failed to update database")
				log.Println(err)
			}

			rowsAffected, err := res.RowsAffected()
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%d rows affected", rowsAffected)
			log.Printf("User %s claimed his score", name)
		}
	}
}

func ws_get_scores(conn *websocket.Conn) {
	// Get scores
	scores, err := getOrderScores()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Database error"))
		log.Println("Failed to fetch scores from database")
		log.Println(err)
		return
	}

	// Marshal scores
	scores_json, err := json.Marshal(scores)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("JSON error"))
		log.Println("Failed to marshal scores")
		log.Println(err)
		return
	}

	conn.WriteMessage(websocket.TextMessage, scores_json)
}
