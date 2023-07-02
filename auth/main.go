package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"

	"google.golang.org/api/option"
)

type SubmissionRequestWithoutAuth struct {
	ProblemID uint `json:"problem_id"`
	Answer    int  `json:"answer"`
}

type SubmissionRequest struct {
	UserID    string `json:"user_id"`
	ProblemID uint   `json:"problem_id"`
	Answer    int    `json:"answer"`
}

type Submission struct {
	ID          uint
	UserID      string
	ProblemID   uint
	Answer      int
	Correct     bool
	ScoreGained int
	CreatedAt   time.Time
}

func main() {
	opt := option.WithCredentialsFile(
		"credentials.json",
	)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	// gin.SetMode(gin.ReleaseMode)
	r.POST("/submission", func(c *gin.Context) {
		var submissionRequestWithoutAuth SubmissionRequestWithoutAuth
		if err := c.ShouldBindJSON(&submissionRequestWithoutAuth); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		idToken := c.GetHeader("Authorization")
		ok := strings.HasPrefix(idToken, "Bearer ")
		if !ok {
			c.AbortWithStatus(403)
			return
		}
		idToken = idToken[7:]

		client, err := app.Auth(c)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		token, err := client.VerifyIDToken(c, idToken)
		if err != nil {
			c.AbortWithStatus(403)
			return
		}

		url := "http://localhost:8080/submission"
		marshalled, err := json.Marshal(SubmissionRequest{
			UserID:    token.UID,
			ProblemID: submissionRequestWithoutAuth.ProblemID,
			Answer:    submissionRequestWithoutAuth.Answer,
		})
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(marshalled))
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatus(resp.StatusCode)
			return
		}

		var submission Submission
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		json.Unmarshal(body, &submission)
		c.JSON(200, submission)
	})

	r.Run(":8081")
}
