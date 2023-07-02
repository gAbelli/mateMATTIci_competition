package main

import "time"

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
	Timestamp string `json:"timestamp"` // only for testing purposes, normally it's empty
}
