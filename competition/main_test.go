package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockSubmissionRequest struct {
	UserId            string
	ProblemID         uint
	Answer            int
	MinutesSinceStart int
	ExpectedScore     int
}

var startTimestamp, _ = time.Parse(time.RFC3339, "2023-07-02T12:00:00.000Z")

// assume there are 3 problems with ids 1, 2, 3 and answers 1, 2, 3
var mockCompetitions [][]*MockSubmissionRequest = [][]*MockSubmissionRequest{
	{
		{"a", 1, 1, 0, 40},
		{"a", 1, 1, 0, 0},
		{"a", 2, 2, 1, 41},
		{"a", 3, 3, 55, 190},
	},
	{
		{"a", 1, 0, 0, -10},
		{"a", 1, 0, 10, -10},
		{"a", 2, 0, 20, -10},
		{"a", 2, 0, 30, -10},
		{"a", 3, 0, 40, -10},
		{"a", 3, 0, 45, -10},
		{"a", 1, 0, 50, -10},
	},
	{
		{"b", 1, 0, 0, -10},
		{"b", 1, 0, 10, -10},
		{"b", 2, 0, 20, -10},
		{"a", 1, 1, 30, 74},
	},
	{
		{"a1", 1, 1, 0, 40},
		{"a1", 2, 2, 0, 40},
		{"a1", 3, 3, 0, 140},
		{"a2", 1, 1, 1, 35},
		{"a2", 2, 2, 1, 35},
		{"a2", 3, 3, 1, 95},
		{"a3", 1, 1, 2, 30},
		{"a3", 2, 2, 2, 30},
		{"a3", 3, 3, 2, 70},
		{"a4", 1, 1, 3, 28},
		{"a4", 2, 2, 3, 28},
		{"a4", 3, 3, 3, 58},
		{"a5", 1, 1, 4, 26},
		{"a5", 2, 2, 4, 26},
		{"a5", 3, 3, 4, 46},
		{"a6", 1, 1, 5, 25},
		{"a6", 2, 2, 5, 25},
		{"a6", 3, 3, 5, 35},
		{"a7", 1, 1, 6, 24},
		{"a7", 2, 2, 6, 24},
		{"a7", 3, 3, 6, 24},
	},
	{
		{"a", 1, 1, 0, 40},
		{"a", 2, 0, 0, -10},
		{"a", 2, 2, 0, 42},
		{"a", 3, 0, 0, -10},
		{"a", 3, 3, 0, 142},
	},
}

func insertMockData() {
	competition := Competition{
		ID:             1234,
		StartTimestamp: startTimestamp,
		EndTimestamp:   startTimestamp.Add(1 * time.Hour),
	}
	problem_1 := Problem{ID: 1, CompetitionID: 1234, Number: 1, CorrectAnswer: 1}
	problem_2 := Problem{ID: 2, CompetitionID: 1234, Number: 2, CorrectAnswer: 2}
	problem_3 := Problem{ID: 3, CompetitionID: 1234, Number: 3, CorrectAnswer: 3}
	db.Create(&competition)
	db.Create(&problem_1)
	db.Create(&problem_2)
	db.Create(&problem_3)
}

func ResetSubmissions() {
	db.Where("problem_id in ?", []uint{1, 2, 3}).Delete(&Submission{})
}

func TestCompetition(t *testing.T) {
	SetupDB()
	r := SetupRouter()
	insertMockData()

	for i, competition := range mockCompetitions {
		ResetSubmissions()
		scores := make(map[string]int)
		for j, mockSubmission := range competition {
			time.Sleep(5 * time.Millisecond)
			mockSubmissionRequest := SubmissionRequest{
				UserID:    mockSubmission.UserId,
				ProblemID: mockSubmission.ProblemID,
				Answer:    mockSubmission.Answer,
				Timestamp: startTimestamp.Add(time.Duration(mockSubmission.MinutesSinceStart) * time.Minute).
					Add(10 * time.Second).
					Format(time.RFC3339),
			}
			marshalled, err := json.Marshal(mockSubmissionRequest)
			if err != nil {
				t.Fatalf("Error in competition %d, submission %d: %v", i, j, err.Error())
			}
			req, err := http.NewRequest("POST", "/submission", bytes.NewReader(marshalled))
			if err != nil {
				panic(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var submission Submission
			json.Unmarshal(w.Body.Bytes(), &submission)
			if submission.ScoreGained != mockSubmission.ExpectedScore {
				t.Fatalf(
					"Error in competition %d, submission %d\nExpected: %d, Got: %d",
					i,
					j,
					mockSubmission.ExpectedScore,
					submission.ScoreGained,
				)
			}
			scores[mockSubmission.UserId] += mockSubmission.ExpectedScore
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("/leaderboard/%d", 1234), nil)
		if err != nil {
			panic(err)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var scoresFromAPI map[string]int
		json.Unmarshal(w.Body.Bytes(), &scoresFromAPI)
		for userID := range scores {
			scores[userID] += 10 * 3
		}
		if len(scores) != len(scoresFromAPI) {
			t.Fatalf(
				"Error in leaderboard, competition %d\nExpected: %d scores, Got: %d scores",
				i,
				len(scores),
				len(scoresFromAPI),
			)
		}
		for userID := range scores {
			if scores[userID] != scoresFromAPI[userID] {
				t.Fatalf(
					"Error in leaderboard, competition %d, user_id %v\nExpected: %d, Got: %d",
					i,
					userID,
					scores[userID],
					scoresFromAPI[userID],
				)
			}
		}
	}

	ResetSubmissions()
}

type MockSubmissionRequestWithError struct {
	UserId            string
	ProblemID         uint
	Answer            int
	MinutesSinceStart int
	ExpectError       bool
}

var mockCompetitionsWithErrors [][]*MockSubmissionRequestWithError = [][]*MockSubmissionRequestWithError{
	{
		{"a", 1, 1, 0, false},
		{"a", 1, 1, -1, true},
		{"a", 1, 1, 61, true},
		{"a", 1, 1, -9999999, true},
	},
}

func TestErrors(t *testing.T) {
	SetupDB()
	r := SetupRouter()
	insertMockData()

	for i, competition := range mockCompetitionsWithErrors {
		ResetSubmissions()
		for j, mockSubmission := range competition {
			time.Sleep(5 * time.Millisecond)
			mockSubmissionRequest := SubmissionRequest{
				UserID:    mockSubmission.UserId,
				ProblemID: mockSubmission.ProblemID,
				Answer:    mockSubmission.Answer,
				Timestamp: startTimestamp.Add(time.Duration(mockSubmission.MinutesSinceStart) * time.Minute).
					Add(10 * time.Second).
					Format(time.RFC3339),
			}
			marshalled, err := json.Marshal(mockSubmissionRequest)
			if err != nil {
				t.Fatalf("Error in competition %d, submission %d: %v", i, j, err.Error())
			}
			req, err := http.NewRequest("POST", "/submission", bytes.NewReader(marshalled))
			if err != nil {
				panic(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			gotError := w.Result().StatusCode != http.StatusOK
			var submission Submission
			json.Unmarshal(w.Body.Bytes(), &submission)
			if gotError != mockSubmission.ExpectError {
				t.Fatalf(
					"Error in competition %d, submission %d\nExpected: %v, Got: %v",
					i,
					j,
					mockSubmission.ExpectError,
					gotError,
				)
			}
		}
	}

	ResetSubmissions()
}
