package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Submission struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID      string    `gorm:"not null"`
	ProblemID   uint      `gorm:"not null"`
	Answer      int       `gorm:"not null"`
	Correct     bool      `gorm:"not null"`
	ScoreGained int       `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Problem     Problem   `gorm:"foreignKey:ProblemID"`
}

type Problem struct {
	ID            uint        `gorm:"primaryKey"`
	CompetitionID uint        `gorm:"not null"`
	CorrectAnswer int         `gorm:"not null"`
	Number        int         `gorm:"not null"`
	Competition   Competition `gorm:"foreignKey:CompetitionID"`
}

type Competition struct {
	ID             uint      `gorm:"primaryKey"`
	StartTimestamp time.Time `gorm:"not null"`
	EndTimestamp   time.Time `gorm:"not null"`
}

type SubmissionRequest struct {
	UserID    string `json:"user_id"`
	ProblemID uint   `json:"problem_id"`
	Answer    int    `json:"answer"`
}

var (
	db       *gorm.DB
	bonusFor []int = []int{
		20, 15, 10, 8, 6, 5, 4, 3, 2, 1,
	}
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func submissionHandler(c *gin.Context) {
	var submissionRequest SubmissionRequest
	if err := c.ShouldBindJSON(&submissionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check that problem_id exist and get the problem
	var problem Problem
	result := db.
		Take(&problem, submissionRequest.ProblemID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "problem does not exist"})
		return
	}

	// get the competition
	var competition Competition
	result = db.
		Take(&competition, problem.CompetitionID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "competition does not exist"})
		return
	}
	if time.Now().After(competition.EndTimestamp) || time.Now().Before(competition.StartTimestamp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "competition is not currently active"})
		return
	}

	// check if the answer is correct
	// if it is wrong, remove 10 points and return
	if problem.CorrectAnswer != submissionRequest.Answer {
		submission := Submission{
			UserID:      submissionRequest.UserID,
			ProblemID:   submissionRequest.ProblemID,
			Answer:      submissionRequest.Answer,
			Correct:     false,
			ScoreGained: -10,
		}
		db.Create(&submission)
		c.JSON(200, submission)
		return
	}

	// check the user had already solved it correctly
	var count int64
	result = db.Model(&Submission{}).Where(
		"user_id = ? AND problem_id = ? AND correct = true",
		submissionRequest.UserID,
		submissionRequest.ProblemID,
	).Distinct("id").Count(&count)
	if count > 0 {
		submission := Submission{
			UserID:      submissionRequest.UserID,
			ProblemID:   submissionRequest.ProblemID,
			Answer:      submissionRequest.Answer,
			Correct:     true,
			ScoreGained: 0,
		}
		c.JSON(200, submission)
		return
	}

	// check how many people have already solved the problem correctly
	var score int
	var firstCorrectSubmission Submission
	result = db.Model(&Submission{}).Where(
		"problem_id = ? AND correct = true",
		submissionRequest.ProblemID,
	).Order("created_at").First(&firstCorrectSubmission)
	until := competition.StartTimestamp.Add(
		5 * competition.EndTimestamp.Sub(competition.StartTimestamp) / 6,
	)
	if result.Error == gorm.ErrRecordNotFound {
		bonus := 20
		// check how many people have submitted a wrong answer to this problem
		// before 5/6 of the total time available
		db.Model(&Submission{}).Where(
			"problem_id = ? AND created_at <= ? AND correct = false",
			submissionRequest.ProblemID,
			until,
		).Distinct("id").Count(&count)
		timeSinceStart := int(
			time.Now().Sub(competition.StartTimestamp).Round(time.Minute).Minutes(),
		)
		score = bonus + 2*int(count) + 1*min(
			timeSinceStart,
			int(until.Sub(competition.StartTimestamp).Round(time.Minute).Minutes()),
		)

		// check if, with this submission, the user solved all problems
		// db.Joins("submissions").Where("competition_id = ? AND user_id = ? AND correct = true")
	} else {
		if firstCorrectSubmission.CreatedAt.Before(until) {
			until = firstCorrectSubmission.CreatedAt
		}

		// count the number of correct submissions before until
		result = db.Model(&Submission{}).Where(
			"problem_id = ? AND correct = true",
			submissionRequest.ProblemID,
		).Distinct("id").Count(&count)
		bonus := 0
		if count <= 10 {
			bonus = bonusFor[count-1]
		}

		// check how many people have submitted a wrong answer to this problem
		// before until
		db.Model(&Submission{}).Where(
			"problem_id = ? AND created_at <= ? AND correct = false",
			submissionRequest.ProblemID,
			until,
		).Distinct("id").Count(&count)
		timeSinceStart := int(
			time.Now().Sub(competition.StartTimestamp).Round(time.Minute).Minutes(),
		)
		score = bonus + 2*int(count) + 1*min(
			timeSinceStart,
			int(until.Sub(competition.StartTimestamp).Round(time.Minute).Minutes()),
		)

		// check if, with this submission, the user solved all problems
	}

	submission := Submission{
		UserID:      submissionRequest.UserID,
		ProblemID:   submissionRequest.ProblemID,
		Answer:      submissionRequest.Answer,
		Correct:     true,
		ScoreGained: score,
	}
	db.Create(&submission)
	c.JSON(200, submission)
}

func main() {
	dsn := "root@tcp(127.0.0.1:3306)/matemattici_competition?charset=utf8mb4&parseTime=True&loc=Local"
	new_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db = new_db

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Submission{}, &Problem{}, &Competition{})

	// for testing purposes, reset the db every time
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

	r := gin.Default()
	r.POST("/", submissionHandler)
	r.Run()
}
