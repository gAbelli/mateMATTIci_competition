package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	Timestamp string `json:"timestamp"` // only for testing purposes, remove in final version
}

var (
	db              *gorm.DB
	bonusForProblem []int = []int{
		20, 15, 10, 8, 6, 5, 4, 3, 2, 1,
	}
	bonusForAllProblems []int = []int{
		100, 60, 40, 30, 20, 10,
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
	now, err := time.Parse(time.RFC3339, submissionRequest.Timestamp)
	if err != nil {
		now = time.Now()
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
	if now.After(competition.EndTimestamp) || now.Before(competition.StartTimestamp) {
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
			CreatedAt:   now, // remove this in production
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

	// check if someone has already solved the problem correctly
	score := 20
	var firstCorrectSubmission Submission
	result = db.Model(&Submission{}).Where(
		"problem_id = ? AND correct = true",
		submissionRequest.ProblemID,
	).Order("created_at").First(&firstCorrectSubmission)
	until := competition.StartTimestamp.Add(
		5 * competition.EndTimestamp.Sub(competition.StartTimestamp) / 6,
	)
	if result.Error == gorm.ErrRecordNotFound {
		bonus := bonusForProblem[0]
		// check how many people have submitted a wrong answer to this problem
		// before 5/6 of the total time available
		db.Model(&Submission{}).Where(
			"problem_id = ? AND created_at <= ? AND correct = false",
			submissionRequest.ProblemID,
			until,
		).Distinct("id").Count(&count)
		timeSinceStart := int(
			now.Sub(competition.StartTimestamp).Round(time.Minute).Minutes(),
		)
		score += bonus + 2*int(count) + 1*min(
			timeSinceStart,
			int(until.Sub(competition.StartTimestamp).Round(time.Minute).Minutes()),
		)
	} else {
		if firstCorrectSubmission.CreatedAt.Before(until) {
			until = firstCorrectSubmission.CreatedAt
		}

		// count the number of correct submissions
		result = db.Model(&Submission{}).Where(
			"problem_id = ? AND correct = true",
			submissionRequest.ProblemID,
		).Distinct("id").Count(&count)
		bonus := 0
		if count <= 9 {
			bonus = bonusForProblem[count]
		}

		// check how many people have submitted a wrong answer to this problem
		// before until
		db.Model(&Submission{}).Where(
			"problem_id = ? AND created_at <= ? AND correct = false",
			submissionRequest.ProblemID,
			until,
		).Distinct("id").Count(&count)
		timeSinceStart := int(
			now.Sub(competition.StartTimestamp).Round(time.Minute).Minutes(),
		)
		score += bonus + 2*int(count) + 1*min(
			timeSinceStart,
			int(until.Sub(competition.StartTimestamp).Round(time.Minute).Minutes()),
		)
	}

	// check if, with this submission, the user solved all problems
	var numberOfProblems int64
	db.Model(Problem{}).
		Where("competition_id = ?", competition.ID).
		Distinct("id").
		Count(&numberOfProblems)

	var solvedAll bool
	db.Raw(
		"select count(problem_id) = ? from submissions join problems on submissions.problem_id = problems.id join competitions on problems.competition_id = competitions.id where competition_id = ? and user_id = ? and problem_id != ? and correct = true",
		numberOfProblems-1,
		competition.ID,
		submissionRequest.UserID,
		problem.ID,
	).Scan(&solvedAll)

	if solvedAll {
		// count how many people have already solved all problems,
		// EXCLUDING the current user
		var peopleWhoSolvedAll int64

		db.Raw(
			"select count(*) from (select (count(problem_id) = ?) as solved_all from submissions join problems on submissions.problem_id = problems.id join competitions on problems.competition_id = competitions.id where competition_id = ? AND correct = true group by submissions.user_id) as derived where derived.solved_all = 1",
			numberOfProblems,
			competition.ID,
		).Scan(&peopleWhoSolvedAll)

		if peopleWhoSolvedAll <= 5 {
			score += bonusForAllProblems[peopleWhoSolvedAll]
		}
	}

	submission := Submission{
		UserID:      submissionRequest.UserID,
		ProblemID:   submissionRequest.ProblemID,
		Answer:      submissionRequest.Answer,
		Correct:     true,
		ScoreGained: score,
		CreatedAt:   now, // remove this in production
	}
	db.Create(&submission)
	c.JSON(200, submission)
}

// func handleLeaderboard(c *gin.Context) {
// 	competitionID, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// }

func SetupDB() {
	dsn := "root@tcp(127.0.0.1:3306)/matemattici_competition?charset=utf8mb4&parseTime=True&loc=Local"
	new_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	db = new_db

	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Submission{}, &Problem{}, &Competition{})
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r.POST("/submission", submissionHandler)
	return r
}

func Reset() {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Submission{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Problem{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Competition{})
}

func main() {
	SetupDB()

	r := SetupRouter()
	r.Run(":8080")
}
