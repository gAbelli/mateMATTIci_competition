package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"
)

func insertMockData() {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Submission{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Problem{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Competition{})

	competition := Competition{
		ID:             1234,
		StartTimestamp: time.Now(),
		EndTimestamp:   time.Now().Add(1 * time.Hour),
	}
	problem_1 := Problem{ID: 1, CompetitionID: 1234, Number: 1, CorrectAnswer: 1}
	problem_2 := Problem{ID: 2, CompetitionID: 1234, Number: 2, CorrectAnswer: 2}
	problem_3 := Problem{ID: 3, CompetitionID: 1234, Number: 3, CorrectAnswer: 3}
	problem_4 := Problem{ID: 4, CompetitionID: 1234, Number: 4, CorrectAnswer: 4}
	problem_5 := Problem{ID: 5, CompetitionID: 1234, Number: 5, CorrectAnswer: 5}
	db.Create(&competition)
	db.Create(&problem_1)
	db.Create(&problem_2)
	db.Create(&problem_3)
	db.Create(&problem_4)
	db.Create(&problem_5)
}

func TestInitialize(t *testing.T) {
	SetupDB()
	Reset()
	insertMockData()
	r := SetupRouter()

	w := httptest.NewRecorder()

	submissionRequest := SubmissionRequest{
		UserID:    "abc",
		ProblemID: 1,
		Answer:    1,
		Timestamp: "",
	}
	marshalled, _ := json.Marshal(submissionRequest)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(marshalled))
	r.ServeHTTP(w, req)
	var submission Submission
	json.Unmarshal(w.Body.Bytes(), &submission)
	fmt.Print(submission)
}
