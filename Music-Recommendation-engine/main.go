package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

type Track struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Genre    string `json:"genre"`
	Duration int    `json:"duration"`
}

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	History []string `json:"history"`
}

type UserItemMatrix map[string]map[string]int

func RecommendationHandler(matrix UserItemMatrix) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		numRecs := 5
		recommendations := GetRecommendations(userID, numRecs, matrix)
		jsonResponse, err := json.Marshal(recommendations)
		if err != nil {
			http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

var tracks = []Track{
	{"4", "Amazing Grace", "Traditional", "Gospel", 180},
	{"1", "Shape of You", "Ed Sheeran", "Pop", 233},
	{"2", "Bohemian Rhapsody", "Queen", "Rock", 354},
	{"3", "Billie Jean", "Michael Jackson", "Pop", 292},
	{"5", "Despacito", "Luis Fonsi", "Latin", 228},
	{"6", "Hotel California", "Eagles", "Rock", 391},
	{"7", "Uptown Funk", "Mark Ronson ft. Bruno Mars", "Funk", 270},
	{"8", "Rolling in the Deep", "Adele", "Soul", 228},
}

var users = []User{
	{"user1", "Tolu", []string{"1", "3"}},
	{"user2", "James", []string{"2"}},
	{"user3", "Ben", []string{"1", "4"}},
	{"user4", "Tom", []string{"2", "3"}},
	{"user5", "Shola", []string{"1", "2"}},
	{"user6", "Paul", []string{"3", "4"}},
	{"user7", "Jacob", []string{"2", "4"}},
}

func ComputeUserItemMatrix(tracks []Track, users []User) UserItemMatrix {
	matrix := make(UserItemMatrix)
	for _, user := range users {
		matrix[user.ID] = make(map[string]int)
		for _, trackID := range user.History {
			matrix[user.ID][trackID]++
		}
	}
	return matrix
}

func CosineSimilarity(a, b map[string]int) float64 {
	dotProduct := 0.0
	magnitudeA := 0.0
	magnitudeB := 0.0
	for key := range a {
		dotProduct += float64(a[key] * b[key])
		magnitudeA += math.Pow(float64(a[key]), 2)
		magnitudeB += math.Pow(float64(b[key]), 2)
	}
	magnitudeA = math.Sqrt(magnitudeA)
	magnitudeB = math.Sqrt(magnitudeB)
	if magnitudeA == 0 || magnitudeB == 0 {
		return 0
	}
	return dotProduct / (magnitudeA * magnitudeB)
}

func GetRecommendations(userID string, numRecs int, matrix UserItemMatrix) []string {
	recommendations := make(map[string]float64)
	userHistory := matrix[userID]
	for _, user := range users {
		if user.ID != userID {
			similarity := CosineSimilarity(userHistory, matrix[user.ID])
			for trackID := range matrix[user.ID] {
				if userHistory[trackID] == 0 {
					recommendations[trackID] += similarity
				}
			}
		}
	}
	sortedRecs := make([]string, 0, len(recommendations))
	for trackID := range recommendations {
		sortedRecs = append(sortedRecs, trackID)
	}

	for i := 0; i < len(sortedRecs)-1; i++ {
		for j := i + 1; j < len(sortedRecs); j++ {
			if recommendations[sortedRecs[i]] < recommendations[sortedRecs[j]] {
				sortedRecs[i], sortedRecs[j] = sortedRecs[j], sortedRecs[i]
			}
		}
	}
	return sortedRecs[:numRecs]
}

func main() {
	matrix := ComputeUserItemMatrix(tracks, users)

	http.HandleFunc("/recommendations", RecommendationHandler(matrix))

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
