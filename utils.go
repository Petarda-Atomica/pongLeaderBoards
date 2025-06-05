package main

import (
	"crypto/rand"
	"database/sql"
	"math/big"
)

type Score struct {
	ID       string         `json:"id"`
	UserName sql.NullString `json:"user_name"`
	Score    int            `json:"score"`
}

func generateID(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func getOrderScores() ([]Score, error) {
	rows, err := scoresDB.Query(`SELECT ID, user_name, score FROM Scores ORDER BY score DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var s Score
		err = rows.Scan(&s.ID, &s.UserName, &s.Score)
		if err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}

	return scores, nil
}
