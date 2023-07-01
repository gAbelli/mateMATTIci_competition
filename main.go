package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Submission struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	CompetitionID string    `gorm:"not null"`
	UserID        string    `gorm:"not null"`
	ProblemID     uint      `gorm:"not null"`
	Answer        int       `gorm:"not null"`
	Correct       bool      `gorm:"not null"`
	ScoreGained   int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

type SubmissionFromUser struct {
	CompetitionID string `json:"competition_id"`
	UserID        string `json:"user_id"`
	ProblemID     uint   `json:"problem_id"`
	Answer        int    `json:"answer"`
}

type Problem struct {
	CompetitionID string `json:"competition_id"`
	ProblemID     uint   `json:"problem_id"`
	CorrectAnswer int    `json:"correct_answer"`
}

var (
	db       *gorm.DB
	problems []*Problem = []*Problem{{CompetitionID: "1", ProblemID: 1, CorrectAnswer: 1}}
)

func submissionHandler(c *gin.Context) {
	var submissionFromUser SubmissionFromUser
	if err := c.ShouldBindJSON(&submissionFromUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check that problem_id exists
	if submissionFromUser.ProblemID != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong id"})
		return
	}

	// check if it is correct
	correct := submissionFromUser.Answer == 1
	if !correct {
		submission := Submission{CompetitionID: submissionFromUser.CompetitionID, UserID: submissionFromUser.UserID, ProblemID: submissionFromUser.ProblemID, Answer: submissionFromUser.Answer, Correct: false, ScoreGained: -10}
		db.Create(&submission)
		c.JSON(200, submission)
		return
	}

	// check if it was already solved correctly
	var submission Submission
	result := db.Where("competition_id = ? AND user_id = ? AND problem_id = ? AND correct = true", submissionFromUser.CompetitionID, submissionFromUser.CompetitionID, submissionFromUser.ProblemID).Find(&submission)
	if result.RowsAffected > 0 {
		submission := Submission{CompetitionID: submissionFromUser.CompetitionID, UserID: submissionFromUser.UserID, ProblemID: submissionFromUser.ProblemID, Answer: submissionFromUser.Answer, Correct: true, ScoreGained: 0}
		c.JSON(200, submission)
		return
	}

	c.JSON(200, submissionFromUser)
}

func main() {
	dsn := "root@tcp(127.0.0.1:3306)/matemattici_competition?charset=utf8mb4&parseTime=True&loc=Local"
	new_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db = new_db

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Submission{})

	// submission := Submission{CompetitionID: "abc", UserID: "abc", ProblemID: 1, Answer: 1, Correct: true, ScoreGained: 0}

	// submission := Submission{CompetitionID: "1", UserID: "abc", ProblemID: 1, Answer: 0, Correct: true, ScoreGained: 20}
	// db.Create(&submission)

	r := gin.Default()
	r.POST("/", submissionHandler)
	r.Run()
}
